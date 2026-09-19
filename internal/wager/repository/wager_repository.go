package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/wager/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var appNamespace = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))

var (
	ErrConflict = errors.New("conflict: idempotency key or external transaction already used with a different payload")
)

type WagerRepository interface {
	Persist(ctx context.Context, wt *domain.WagerTransaction, msgId, consumerName *string) (PersistResult, error)
}

type PersistResult struct {
	TransactionID    string
	Status           domain.TransactionState
	BalanceAmount    string
	BalanceCurrency  string
	IdempotentReplay bool
}

type postgresWagerRepository struct {
	db *pgxpool.Pool
}

func NewPostgresWagerRepository(db *pgxpool.Pool) WagerRepository {
	return &postgresWagerRepository{db: db}
}

func (r *postgresWagerRepository) Persist(ctx context.Context, wt *domain.WagerTransaction, msgId, consumerName *string) (PersistResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return PersistResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if msgId != nil && consumerName != nil {
		var exists bool
		checkInboxQuery := `SELECT EXISTS(SELECT 1 FROM inbox WHERE consumer_name = $1 AND message_id = $2)`
		err = tx.QueryRow(ctx, checkInboxQuery, *consumerName, *msgId).Scan(&exists)
		if err != nil {
			return PersistResult{}, fmt.Errorf("failed to check inbox: %w", err)
		}
		if exists {
			var existingID string
			var existingStatus domain.TransactionState
			var existingBalanceAmount int64
			var existingBalanceCurrency string
			queryReplay := `
				SELECT wt.id, wt.status, w.balance_amount, w.currency
				FROM wager_transactions wt
				JOIN wallets w ON w.id = wt.wallet_id
				WHERE wt.provider_id = $1 AND wt.external_transaction_id = $2
				LIMIT 1
			`
			var providerVal any
			if wt.Provider() != nil {
				providerVal = *wt.Provider()
			}
			var extIDVal any
			if wt.ExternalID() != nil {
				extIDVal = *wt.ExternalID()
			}
			err = tx.QueryRow(ctx, queryReplay, providerVal, extIDVal).Scan(
				&existingID, &existingStatus, &existingBalanceAmount, &existingBalanceCurrency,
			)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					var fallbackBalance int64
					var fallbackCurrency string
					_ = tx.QueryRow(ctx, `SELECT balance_amount, currency FROM wallets WHERE id = $1`, wt.WalletID()).Scan(&fallbackBalance, &fallbackCurrency)

					balanceStr := fmt.Sprintf("%.2f", float64(fallbackBalance)/100)
					return PersistResult{
						Status:           domain.StateProcessed,
						BalanceAmount:    balanceStr,
						BalanceCurrency:  fallbackCurrency,
						IdempotentReplay: true,
					}, nil
				}
				return PersistResult{}, fmt.Errorf("failed to fetch replayed transaction from inbox: %w", err)
			}
			balanceStr := fmt.Sprintf("%.2f", float64(existingBalanceAmount)/100)
			return PersistResult{
				TransactionID:    existingID,
				Status:           existingStatus,
				BalanceAmount:    balanceStr,
				BalanceCurrency:  existingBalanceCurrency,
				IdempotentReplay: true,
			}, nil
		}
		insertInboxQuery := `INSERT INTO inbox (consumer_name, message_id, processed_at, payload_hash) VALUES ($1, $2, NOW(), $3)`
		_, err = tx.Exec(ctx, insertInboxQuery, *consumerName, *msgId, wt.PayloadHash())
		if err != nil {
			return PersistResult{}, fmt.Errorf("failed to insert into inbox: %w", err)
		}
	}
	var (
		existingID              string
		existingPayloadHash     string
		existingStatus          domain.TransactionState
		existingBalanceAmount   int64
		existingBalanceCurrency string
	)
	queryCheck := `
		SELECT wt.id, wt.payload_hash, wt.status, w.balance_amount, w.currency
		FROM wager_transactions wt
		JOIN wallets w ON w.id = wt.wallet_id
		WHERE ($1::text IS NOT NULL AND wt.idempotency_key = $1)
		OR ($2::text IS NOT NULL AND $3::text IS NOT NULL AND wt.provider_id = $2 AND wt.external_transaction_id = $3)
		LIMIT 1;
	`
	var idempotencyKeyVal, providerVal, externalIDVal any
	if wt.IdempotencyKey() != nil {
		idempotencyKeyVal = *wt.IdempotencyKey()
	}
	if wt.Provider() != nil {
		providerVal = *wt.Provider()
	}
	if wt.ExternalID() != nil {
		externalIDVal = *wt.ExternalID()
	}
	err = tx.QueryRow(ctx, queryCheck, idempotencyKeyVal, providerVal, externalIDVal).Scan(
		&existingID,
		&existingPayloadHash,
		&existingStatus,
		&existingBalanceAmount,
		&existingBalanceCurrency,
	)
	if err == nil {
		var incomingHash string
		if wt.PayloadHash() != nil {
			incomingHash = *wt.PayloadHash()
		}
		if existingPayloadHash != incomingHash {
			return PersistResult{}, ErrConflict
		}
		balanceStr := fmt.Sprintf("%.2f", float64(existingBalanceAmount)/100)
		return PersistResult{
			TransactionID:    existingID,
			Status:           existingStatus,
			BalanceAmount:    balanceStr,
			BalanceCurrency:  existingBalanceCurrency,
			IdempotentReplay: true,
		}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PersistResult{}, fmt.Errorf("failed to check existing transaction: %w", err)
	}
	isReversal := wt.Type() == domain.TypeRefund || wt.Type() == domain.TypeRollback
	var refID string
	var refStatus domain.TransactionState
	var refAmount int64
	var refPlayerID, refWalletID, refCurrency, refKind string
	if isReversal {
		refExtID := wt.ReferenceExternalTransactionID()
		if refExtID == nil {
			return PersistResult{}, errors.New("referenceExternalTransactionId is required for refunds and rollbacks")
		}
		queryRef := `
			SELECT id, status, amount_value, player_id, wallet_id, amount_currency, kind
			FROM wager_transactions
			WHERE provider_id = $1 AND external_transaction_id = $2
			LIMIT 1;
		`
		err = tx.QueryRow(ctx, queryRef, providerVal, *refExtID).Scan(
			&refID, &refStatus, &refAmount, &refPlayerID, &refWalletID, &refCurrency, &refKind,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			wt.TransitionTo(domain.StatePendingReference, nil, time.Now())
		} else if err != nil {
			return PersistResult{}, fmt.Errorf("failed to check reference transaction: %w", err)
		} else {
			if refStatus != domain.StateProcessed {
				return PersistResult{}, errors.New("referenced transaction is not in PROCESSED state")
			}
			if refPlayerID != wt.PlayerID() || refWalletID != wt.WalletID() {
				return PersistResult{}, errors.New("reference mismatch: player or wallet differs")
			}
			if refAmount != wt.Amount().Amount() || refCurrency != wt.Amount().Currency() {
				return PersistResult{}, errors.New("reference mismatch: amount or currency differs (partial refunds not allowed)")
			}
			var countReversals int
			queryCheckRev := `
				SELECT COUNT(1) FROM wager_transactions 
				WHERE provider_id = $1 AND reference_external_transaction_id = $2 
				  AND kind = $3 AND status = 'PROCESSED'
			`
			_ = tx.QueryRow(ctx, queryCheckRev, providerVal, *refExtID, string(wt.Type())).Scan(&countReversals)
			if countReversals > 0 {
				return PersistResult{}, errors.New("transaction already fully refunded/rolled back")
			}
		}
	}
	var currentBalance int64
	var walletCurrency string
	var walletVersion int64
	lockWalletQuery := `
		SELECT balance_amount, currency, version
		FROM wallets
		WHERE id = $1
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockWalletQuery, wt.WalletID()).Scan(&currentBalance, &walletCurrency, &walletVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PersistResult{}, errors.New("wallet not found")
		}
		return PersistResult{}, fmt.Errorf("failed to lock wallet: %w", err)
	}
	txAmountInt := wt.Amount().Amount()
	txCurrency := wt.Amount().Currency()
	if walletCurrency != txCurrency {
		return PersistResult{}, errors.New("currency mismatch between wallet and transaction")
	}
	var newBalance int64 = currentBalance
	var direction string
	if wt.State() != domain.StatePendingReference {
		switch wt.Type() {
		case domain.TypeBet:
			direction = "DEBIT"
			if currentBalance < txAmountInt {
				return PersistResult{}, errors.New("insufficient funds")
			}
			newBalance = currentBalance - txAmountInt
		case domain.TypeWin, domain.TypeRefund:
			direction = "CREDIT"
			newBalance = currentBalance + txAmountInt
		case domain.TypeRollback:
			if refKind == string(domain.TypeBet) {
				direction = "CREDIT"
				newBalance = currentBalance + txAmountInt
			} else if refKind == string(domain.TypeWin) || refKind == string(domain.TypeRefund) {
				direction = "DEBIT"
				if currentBalance < txAmountInt {
					return PersistResult{}, errors.New("insufficient funds for rollback")
				}
				newBalance = currentBalance - txAmountInt
			} else {
				return PersistResult{}, fmt.Errorf("unsupported reference kind for rollback: %s", refKind)
			}
		case domain.TypeLoss:
			direction = ""
			newBalance = currentBalance
		case domain.TypeOpening:
			direction = "CREDIT"
			newBalance = currentBalance + txAmountInt
		default:
			return PersistResult{}, fmt.Errorf("unsupported transaction type: %v", wt.Type())
		}
	}
	insertTxQuery := `
		INSERT INTO wager_transactions (
			id, provider_id, external_transaction_id, idempotency_key, payload_hash,
			wallet_id, player_id, round_id, game_id, kind, amount_value, amount_currency, 
			reference_external_transaction_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15);
	`
	var refExtIDVal any
	if isReversal {
		refExtIDVal = wt.ReferenceExternalTransactionID()
	}
	_, err = tx.Exec(ctx, insertTxQuery,
		wt.ID(),
		wt.Provider(),
		wt.ExternalID(),
		wt.IdempotencyKey(),
		wt.PayloadHash(),
		wt.WalletID(),
		wt.PlayerID(),
		wt.RoundID(),
		wt.GameID(),
		string(wt.Type()),
		txAmountInt,
		txCurrency,
		refExtIDVal,
		string(wt.State()),
		wt.CreatedAt(),
	)
	if err != nil {
		return PersistResult{}, fmt.Errorf("failed to insert wager transaction: %w", err)
	}
	if wt.State() != domain.StatePendingReference && direction != "" {
		updateWalletQuery := `
			UPDATE wallets
			SET balance_amount = $1, version = version + 1, updated_at = NOW()
			WHERE id = $2 AND version = $3;
		`
		res, err := tx.Exec(ctx, updateWalletQuery, newBalance, wt.WalletID(), walletVersion)
		if err != nil {
			return PersistResult{}, fmt.Errorf("failed to update wallet balance: %w", err)
		}
		if res.RowsAffected() == 0 {
			return PersistResult{}, errors.New("concurrency conflict updating wallet")
		}

		insertLedgerQuery := `
			INSERT INTO wallet_ledger_entries (
				wallet_id, transaction_id, direction, amount_value, amount_currency,
				balance_before_value, balance_before_currency, balance_after_value, balance_after_currency, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
		`
		_, err = tx.Exec(ctx, insertLedgerQuery,
			wt.WalletID(),
			wt.ID(),
			direction,
			txAmountInt,
			txCurrency,
			currentBalance,
			walletCurrency,
			newBalance,
			walletCurrency,
			wt.CreatedAt(),
		)
		if err != nil {
			return PersistResult{}, fmt.Errorf("failed to insert ledger entry: %w", err)
		}
	}
	insertOutboxQuery := `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, occurred_at, event_id, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	var correlationID any
	if *wt.RoundID() != "" {
		correlationID = uuid.NewSHA1(appNamespace, []byte(*wt.RoundID())).String()
	} else {
		correlationID = uuid.NewSHA1(appNamespace, []byte(wt.ID())).String()
	}
	eventType := "WagerTransactionProcessed"
	if wt.State() == domain.StateRejected {
		eventType = "WagerTransactionRejected"
	} else if wt.State() == domain.StatePendingReference {
		eventType = "WagerTransactionPendingReference"
	}
	eventId := uuid.NewSHA1(appNamespace, []byte(*wt.IdempotencyKey())).String()
	eventPayload := fmt.Sprintf(`{"transactionId": "%s", "walletId": "%s", "status": "%s"}`, wt.ID(), wt.WalletID(), wt.State())
	_, err = tx.Exec(ctx, insertOutboxQuery,
		"WagerTransaction",
		wt.ID(),
		eventType,
		eventPayload,
		wt.CreatedAt(),
		eventId,
		correlationID,
	)
	if err != nil {
		return PersistResult{}, fmt.Errorf("failed to insert outbox event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return PersistResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	newBalanceStr := fmt.Sprintf("%.2f", float64(newBalance)/100)
	return PersistResult{
		TransactionID:    wt.ID(),
		Status:           wt.State(),
		BalanceAmount:    newBalanceStr,
		BalanceCurrency:  walletCurrency,
		IdempotentReplay: false,
	}, nil
}

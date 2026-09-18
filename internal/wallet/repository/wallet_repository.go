package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/FelipeSoft/backend-challenge-go/internal/wallet/domain/wallet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var appNamespace = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))

type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet domain.Wallet, wagerTransactionId string, payloadHash string) (string, error)
}

type postgresWalletRepo struct {
	db *pgxpool.Pool
}

func NewPostgresWalletRepository(db *pgxpool.Pool) WalletRepository {
	return &postgresWalletRepo{db: db}
}

func (r *postgresWalletRepo) CreateWallet(ctx context.Context, wallet domain.Wallet, wagerTransactionId string, payloadHash string) (string, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("fail to begin SQL transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)
	walletID := wallet.ID()
	playerID := wallet.PlayerID()
	currency := wallet.Currency()
	amount := wallet.Balance().Amount()
	now := time.Now()
	queryWallet := `
		INSERT INTO wallets (id, player_id, currency, balance_amount, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 1, $5, $6);
	`
	_, err = dbTx.Exec(ctx, queryWallet, walletID, playerID, currency, amount, now, now)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	if amount > 0 {
		queryTx := `
			INSERT INTO wager_transactions (
				id, wallet_id, player_id, kind, amount_value, amount_currency, 
				status, payload_hash, created_at, updated_at
			)
			VALUES ($1, $2, $3, 'OPENING', $4, $5, 'PROCESSED', $6, $7, $8);
		`
		_, err = dbTx.Exec(ctx, queryTx, wagerTransactionId, walletID, playerID, amount, currency, payloadHash, now, now)
		if err != nil {
			return "", fmt.Errorf("error to insert wager transaction OPENING: %w", err)
		}
		queryLedger := `
			INSERT INTO wallet_ledger_entries (
				wallet_id, transaction_id, direction, amount_value, amount_currency, 
				balance_before_value, balance_before_currency, balance_after_value, balance_after_currency, created_at
			)
			VALUES ($1, $2, 'CREDIT', $3, $4, 0, $5, $3, $4, $6);
		`
		_, err = dbTx.Exec(ctx, queryLedger, walletID, wagerTransactionId, amount, currency, currency, now)
		if err != nil {
			return "", fmt.Errorf("error to insert ledger entry: %w", err)
		}
		processedEventPayload, _ := json.Marshal(map[string]interface{}{
			"transactionId": wagerTransactionId,
			"walletId":      walletID,
			"status":        "PROCESSED",
		})
		balanceChangedPayload, _ := json.Marshal(map[string]interface{}{
			"walletId":      walletID,
			"transactionId": wagerTransactionId,
			"direction":     "CREDIT",
			"balanceBefore": 0,
			"balanceAfter":  amount,
			"walletVersion": 1,
		})
		wagerTransactionEventId := uuid.NewSHA1(appNamespace, []byte(wagerTransactionId+"-WagerTransactionProcessed")).String()
		walletEventId := uuid.NewSHA1(appNamespace, []byte(wagerTransactionId+"-WalletBalanceChanged")).String()
		queryOutbox := `
            INSERT INTO outbox (event_id, correlation_id, aggregate_type, aggregate_id, event_type, payload, occurred_at)
            VALUES 
                ($5, $7, 'WagerTransaction', $1, 'WagerTransactionProcessed', $2, $4),
                ($6, $7, 'Wallet', $1, 'WalletBalanceChanged', $3, $4);
        `
		_, err = dbTx.Exec(ctx, queryOutbox,
			walletID,
			processedEventPayload,
			balanceChangedPayload,
			now,
			wagerTransactionEventId,
			walletEventId,
			wagerTransactionId,
		)
		if err != nil {
			return "", fmt.Errorf("error to insert outbox lines: %w", err)
		}
	}
	if err := dbTx.Commit(ctx); err != nil {
		return "", fmt.Errorf("fail to commit opening transaction: %w", err)
	}
	return walletID, nil
}

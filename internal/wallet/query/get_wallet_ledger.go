package query

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerEntryResponse struct {
	ID            string       `json:"id"`
	TransactionID string       `json:"transactionId"`
	Direction     string       `json:"direction"`
	Amount        domain.Money `json:"amount"`
	BalanceBefore domain.Money `json:"balanceBefore"`
	BalanceAfter  domain.Money `json:"balanceAfter"`
	CreatedAt     time.Time    `json:"createdAt"`
}

type PaginationMeta struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}

type GetWalletLedgerResponse struct {
	Data       []LedgerEntryResponse `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

type GetWalletLedger struct {
	db *pgxpool.Pool
}

func NewGetWalletLedger(db *pgxpool.Pool) *GetWalletLedger {
	return &GetWalletLedger{db: db}
}

func (s *GetWalletLedger) Execute(ctx context.Context, walletId string, cursor string, limitStr string) (GetWalletLedgerResponse, error) {
	if walletId == "" {
		return GetWalletLedgerResponse{}, fmt.Errorf("walletId cannot be empty")
	}
	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	query := `
		SELECT 
			id, 
			transaction_id, 
			direction, 
			amount_value, 
			amount_currency, 
			balance_before_value, 
			balance_before_currency, 
			balance_after_value, 
			balance_after_currency, 
			created_at
		FROM wallet_ledger_entries
		WHERE wallet_id = $1
	`
	args := []any{walletId}
    argIdx := 2
    if cursor != "" {
        query += fmt.Sprintf(" AND (created_at, id) < (SELECT created_at, id FROM wallet_ledger_entries WHERE id = $%d)", argIdx)
        args = append(args, cursor)
        argIdx++
    }
    query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIdx)
    args = append(args, limit+1)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return GetWalletLedgerResponse{}, fmt.Errorf("falha ao consultar ledger da carteira: %w", err)
	}
	defer rows.Close()
	var entries []LedgerEntryResponse
	for rows.Next() {
		var e LedgerEntryResponse
		var amountVal, balanceBeforeVal, balanceAfterVal int64
		var amountCurr, balanceBeforeCurr, balanceAfterCurr string
		err := rows.Scan(
			&e.ID,
			&e.TransactionID,
			&e.Direction,
			&amountVal,
			&amountCurr,
			&balanceBeforeVal,
			&balanceBeforeCurr,
			&balanceAfterVal,
			&balanceAfterCurr,
			&e.CreatedAt,
		)
		if err != nil {
			return GetWalletLedgerResponse{}, fmt.Errorf("falha ao escanear entrada do ledger: %w", err)
		}
		amountMoney, err := domain.NewMoneyFromInt(amountVal, amountCurr)
		if err != nil {
			return GetWalletLedgerResponse{}, fmt.Errorf("erro ao criar money para amount: %w", err)
		}
		balanceBeforeMoney, err := domain.NewMoneyFromInt(balanceBeforeVal, balanceBeforeCurr)
		if err != nil {
			return GetWalletLedgerResponse{}, fmt.Errorf("erro ao criar money para balanceBefore: %w", err)
		}
		balanceAfterMoney, err := domain.NewMoneyFromInt(balanceAfterVal, balanceAfterCurr)
		if err != nil {
			return GetWalletLedgerResponse{}, fmt.Errorf("erro ao criar money para balanceAfter: %w", err)
		}
		e.Amount = amountMoney
		e.BalanceBefore = balanceBeforeMoney
		e.BalanceAfter = balanceAfterMoney
		entries = append(entries, e)
	}
	hasMore := false
	var nextCursor string
	if len(entries) > limit {
		hasMore = true
		entries = entries[:limit]
		nextCursor = entries[len(entries)-1].ID
	}
	return GetWalletLedgerResponse{
		Data: entries,
		Pagination: PaginationMeta{
			NextCursor: nextCursor,
			Limit:      limit,
			HasMore:    hasMore,
		},
	}, nil
}
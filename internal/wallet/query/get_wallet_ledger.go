package query

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerEntryResponse struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transactionId"`
	Direction     string    `json:"direction"`
	Amount        MoneyDTO  `json:"amount"`
	BalanceBefore MoneyDTO  `json:"balanceBefore"`
	BalanceAfter  MoneyDTO  `json:"balanceAfter"`
	CreatedAt     time.Time `json:"createdAt"`
}

type GetWalletLedgerResponse struct {
	Data       []LedgerEntryResponse `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

type PaginationMeta struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
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
			limit = min(parsed, 100)
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
		query += fmt.Sprintf(" AND id < $%d", argIdx)
		args = append(args, cursor)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIdx)
	args = append(args, limit+1)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return GetWalletLedgerResponse{}, fmt.Errorf("falha ao consultar ledger: %w", err)
	}
	defer rows.Close()
	var entries []LedgerEntryResponse
	for rows.Next() {
		var e LedgerEntryResponse
		err := rows.Scan(
			&e.ID,
			&e.TransactionID,
			&e.Direction,
			&e.Amount.Amount,
			&e.Amount.Currency,
			&e.BalanceBefore.Amount,
			&e.BalanceBefore.Currency,
			&e.BalanceAfter.Amount,
			&e.BalanceAfter.Currency,
			&e.CreatedAt,
		)
		if err != nil {
			return GetWalletLedgerResponse{}, fmt.Errorf("falha ao escanear ledger: %w", err)
		}
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
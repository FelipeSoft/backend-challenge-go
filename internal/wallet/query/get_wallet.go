package query

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetWalletResponse struct {
	ID        string       `json:"id"`
	PlayerID  string       `json:"playerID"`
	Balance   domain.Money `json:"balance" binding:"required"`
	Version   int64        `json:"version"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type GetWallet struct {
	db *pgxpool.Pool
}

func NewGetWallet(db *pgxpool.Pool) *GetWallet {
	return &GetWallet{
		db: db,
	}
}

func (s *GetWallet) Execute(ctx context.Context, walletId string) (GetWalletResponse, error) {
	if walletId == "" {
		return GetWalletResponse{}, fmt.Errorf("walletId cannot be empty")
	}
	query := `
		SELECT id, player_id, currency, balance_amount, version, created_at, updated_at 
		FROM wallets 
		WHERE id = $1
	`
	var (
		id, playerID, currency string
		balanceAmount, version int64
		createdAt, updatedAt   time.Time
	)
	err := s.db.QueryRow(ctx, query, walletId).Scan(
		&id,
		&playerID,
		&currency,
		&balanceAmount,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return GetWalletResponse{}, fmt.Errorf("wallet not found: %w", err)
	}
	money, err := domain.NewMoneyFromInt(balanceAmount, currency)
	if err != nil {
		return GetWalletResponse{}, fmt.Errorf("error creating Money: %w", err)
	}
	return GetWalletResponse{
		ID:        id,
		PlayerID:  playerID,
		Balance:   money,
		Version:   version,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

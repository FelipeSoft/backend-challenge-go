package repository

import (
	"context"
	"fmt"

	domain "github.com/FelipeSoft/jungle-gaming/internal/wallet/domain/wallet"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet domain.Wallet) (string, error)
}

type postgresWalletRepo struct {
	db *pgxpool.Pool
}

func NewPostgresWalletRepository(db *pgxpool.Pool) WalletRepository {
	return &postgresWalletRepo{db: db}
}

func (r *postgresWalletRepo) CreateWallet(ctx context.Context, wallet domain.Wallet) (string, error) {
	var walletIDCreated string
	query := `
		INSERT INTO wallets (
			id, 
			player_id, 
			currency, 
			balance_amount, 
			version, 
			created_at, 
			updated_at
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id;
	`
	err := r.db.QueryRow(
		ctx, 
		query, 
		wallet.ID(),
		wallet.PlayerID(),
		wallet.Currency(),
		wallet.Balance().Amount(),
		wallet.Version(),
		wallet.CreatedAt(),
		wallet.UpdatedAt(),
	).Scan(&walletIDCreated)
	if err != nil {
		return "", fmt.Errorf("erro ao criar carteira: %w", err)
	}
	return walletIDCreated, nil
}
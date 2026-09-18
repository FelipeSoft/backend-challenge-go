package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetProviderWagerTransactionResponse struct {
	ID                           string       `json:"id"`
	ProviderID                   *string      `json:"providerId,omitempty"`
	ExternalTransactionID        *string      `json:"externalTransactionId,omitempty"`
	WalletID                     string       `json:"walletId"`
	PlayerID                     string       `json:"playerId"`
	RoundID                      *string      `json:"roundId,omitempty"`
	GameID                       *string      `json:"gameId,omitempty"`
	Kind                         string       `json:"kind"`
	Amount                       domain.Money `json:"amount"`
	ReferenceExternalTransactionId *string    `json:"referenceExternalTransactionId,omitempty"`
	Status                       string       `json:"status"`
	FailureCode                  *string      `json:"failureCode,omitempty"`
	ResultAmount                 *domain.Money `json:"resultAmount,omitempty"`
	CreatedAt                    time.Time    `json:"createdAt"`
	UpdatedAt                    time.Time    `json:"updatedAt"`
}

type GetProviderWagerTransaction struct {
	db *pgxpool.Pool
}

func NewGetProviderWagerTransaction(db *pgxpool.Pool) *GetProviderWagerTransaction {
	return &GetProviderWagerTransaction{
		db: db,
	}
}

func (s *GetProviderWagerTransaction) Execute(
	ctx context.Context,
	externalTransactionId string,
	providerId string,
) (GetProviderWagerTransactionResponse, error) {
	if externalTransactionId == "" || providerId == "" {
		return GetProviderWagerTransactionResponse{}, fmt.Errorf("externalTransactionId and providerId cannot be empty")
	}
	query := `
		SELECT 
			id, 
			provider_id, 
			external_transaction_id, 
			wallet_id, 
			player_id, 
			round_id, 
			game_id, 
			kind, 
			amount_value, 
			amount_currency, 
			reference_external_transaction_id, 
			status, 
			failure_code, 
			result_amount_value, 
			result_amount_currency, 
			created_at, 
			updated_at 
		FROM wager_transactions 
		WHERE provider_id = $1 AND external_transaction_id = $2
	`
	var (
		id, walletID, playerID, kind, status string
		provID, extTxID, roundID, gameID, refExtTxID, failureCode *string
		amountVal, resultAmountVal *int64
		amountCurr, resultAmountCurr *string
		createdAt, updatedAt time.Time
	)
	err := s.db.QueryRow(ctx, query, providerId, externalTransactionId).Scan(
		&id,
		&provID,
		&extTxID,
		&walletID,
		&playerID,
		&roundID,
		&gameID,
		&kind,
		&amountVal,
		&amountCurr,
		&refExtTxID,
		&status,
		&failureCode,
		&resultAmountVal,
		&resultAmountCurr,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return GetProviderWagerTransactionResponse{}, fmt.Errorf("wager transaction not found: %w", err)
	}
	amountMoney, err := domain.NewMoneyFromInt(*amountVal, *amountCurr)
	if err != nil {
		return GetProviderWagerTransactionResponse{}, fmt.Errorf("error creating Money for amount: %w", err)
	}
	var resultMoney *domain.Money
	if resultAmountVal != nil && resultAmountCurr != nil {
		money, err := domain.NewMoneyFromInt(*resultAmountVal, *resultAmountCurr)
		if err != nil {
			return GetProviderWagerTransactionResponse{}, fmt.Errorf("error creating Money for result amount: %w", err)
		}
		resultMoney = &money
	}
	return GetProviderWagerTransactionResponse{
		ID:                           id,
		ProviderID:                   provID,
		ExternalTransactionID:        extTxID,
		WalletID:                     walletID,
		PlayerID:                     playerID,
		RoundID:                      roundID,
		GameID:                       gameID,
		Kind:                         kind,
		Amount:                       amountMoney,
		ReferenceExternalTransactionId: refExtTxID,
		Status:                       status,
		FailureCode:                  failureCode,
		ResultAmount:                 resultMoney,
		CreatedAt:                    createdAt,
		UpdatedAt:                    updatedAt,
	}, nil
}
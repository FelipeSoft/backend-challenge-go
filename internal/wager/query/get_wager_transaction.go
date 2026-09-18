package query

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetWagerTransactionResponse struct {
	TransactionID                  string        `json:"transactionId"`
	ProviderID                     *string       `json:"providerId,omitempty"`
	ExternalTransactionID          *string       `json:"externalTransactionId,omitempty"`
	IdempotencyKey                 *string       `json:"idempotencyKey,omitempty"`
	WalletID                       string        `json:"walletId"`
	PlayerID                       string        `json:"playerId"`
	RoundID                        *string       `json:"roundId,omitempty"`
	GameID                         *string       `json:"gameId,omitempty"`
	Kind                           string        `json:"kind"`
	Amount                         domain.Money  `json:"amount"`
	ReferenceExternalTransactionID *string       `json:"referenceExternalTransactionId,omitempty"`
	Status                         string        `json:"status"`
	FailureCode                    *string       `json:"failureCode,omitempty"`
	ResultAmount                   *domain.Money `json:"resultAmount,omitempty"`
	CreatedAt                      time.Time     `json:"createdAt"`
	UpdatedAt                      time.Time     `json:"updatedAt"`
}

type GetWagerTransaction struct {
	db *pgxpool.Pool
}

func NewGetWagerTransaction(db *pgxpool.Pool) *GetWagerTransaction {
	if db == nil {
        panic("NewGetWagerTransaction: db pool cannot be nil")
    }
	return &GetWagerTransaction{db: db}
}

func (s *GetWagerTransaction) Execute(ctx context.Context, transactionId string) (GetWagerTransactionResponse, error) {
	if transactionId == "" {
		return GetWagerTransactionResponse{}, fmt.Errorf("transactionId cannot be empty")
	}
	query := `
		SELECT 
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
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
		WHERE id = $1
	`
	var (
		id                             string
		providerID                     *string
		externalTransactionID          *string
		idempotencyKey                 *string
		walletID                       string
		playerID                       string
		roundID                        *string
		gameID                         *string
		kind                           string
		amountValue                    int64
		amountCurrency                 string
		referenceExternalTransactionID *string
		status                         string
		failureCode                    *string
		resultAmountValue              *int64
		resultAmountCurrency           *string
		createdAt                      time.Time
		updatedAt                      time.Time
	)
	err := s.db.QueryRow(ctx, query, transactionId).Scan(
		&id,
		&providerID,
		&externalTransactionID,
		&idempotencyKey,
		&walletID,
		&playerID,
		&roundID,
		&gameID,
		&kind,
		&amountValue,
		&amountCurrency,
		&referenceExternalTransactionID,
		&status,
		&failureCode,
		&resultAmountValue,
		&resultAmountCurrency,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return GetWagerTransactionResponse{}, fmt.Errorf("transação não encontrada: %w", err)
	}
	amountMoney, err := domain.NewMoneyFromInt(amountValue, amountCurrency)
	if err != nil {
		return GetWagerTransactionResponse{}, fmt.Errorf("erro ao criar money para amount: %w", err)
	}
	var resultAmount *domain.Money
	if resultAmountValue != nil && resultAmountCurrency != nil {
		rm, err := domain.NewMoneyFromInt(*resultAmountValue, *resultAmountCurrency)
		if err != nil {
			return GetWagerTransactionResponse{}, fmt.Errorf("erro ao criar money para resultAmount: %w", err)
		}
		resultAmount = &rm
	}
	return GetWagerTransactionResponse{
		TransactionID:                  id,
		ProviderID:                     providerID,
		ExternalTransactionID:          externalTransactionID,
		IdempotencyKey:                 idempotencyKey,
		WalletID:                       walletID,
		PlayerID:                       playerID,
		RoundID:                        roundID,
		GameID:                         gameID,
		Kind:                           kind,
		Amount:                         amountMoney,
		ReferenceExternalTransactionID: referenceExternalTransactionID,
		Status:                         status,
		FailureCode:                    failureCode,
		ResultAmount:                   resultAmount,
		CreatedAt:                      createdAt,
		UpdatedAt:                      updatedAt,
	}, nil
}
package usecase

import (
	"context"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	wager "github.com/FelipeSoft/backend-challenge-go/internal/wager/domain"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/repository"
	"github.com/google/uuid"
)

type CreateWagerTransaction struct {
	wagerRepository repository.WagerRepository
}

type CreateWagerTransactionInput struct {
    ProviderID                     string  `json:"provider_id"`
    ExternalTransactionID          string  `json:"external_transaction_id"`
    PlayerID                       string  `json:"player_id"`
    WalletID                       string  `json:"wallet_id"`
    RoundID                        *string `json:"round_id"`
    GameID                         *string `json:"game_id"`
    Kind                           string  `json:"kind"`
    MoneyAmount                    string  `json:"money_amount"`
    MoneyCurrency                  string  `json:"money_currency"`
    IdempotencyKey                 string  `json:"idempotency_key"`
    ReferenceExternalTransactionId *string `json:"reference_external_transaction_id"`
    MessageID                      *string `json:"message_id"`
    ConsumerName                   *string `json:"consumer_name"`
    PayloadHash                    *string `json:"payload_hash"`
}

type CreateWagerTransactionOutput struct {
	TransactionID    string       `json:"transactionId"`
	Status           string       `json:"status"`
	Balance          domain.Money `json:"balance"`
	IdempotentReplay bool         `json:"idempotentReplay"`
}

func NewCreateWagerTransaction(wagerRepository repository.WagerRepository) *CreateWagerTransaction {
	return &CreateWagerTransaction{
		wagerRepository: wagerRepository,
	}
}

func (uc *CreateWagerTransaction) Execute(ctx context.Context, input CreateWagerTransactionInput) (CreateWagerTransactionOutput, error) {
	money, err := domain.NewMoneyFromString(input.MoneyAmount, input.MoneyCurrency)
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	txType := wager.TransactionType(input.Kind)
	now := time.Now()
	internalID, err := uuid.NewV7()
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	payloadHash, err := wager.ComputePayloadHash(
		input.ProviderID,
		input.ExternalTransactionID,
		input.PlayerID,
		input.WalletID,
		input.Kind,
		input.RoundID,
		input.GameID,
		input.ReferenceExternalTransactionId,
		money.Amount(),
		money.Currency(),
	)
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	wagerTx, err := wager.NewExternalTransaction(
		internalID.String(),
		input.ExternalTransactionID,
		input.ProviderID,
		input.IdempotencyKey,
		payloadHash,
		input.WalletID,
		input.PlayerID,
		input.RoundID,
		input.GameID,
		txType,
		money,
		input.ReferenceExternalTransactionId,
		now,
	)
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	result, err := uc.wagerRepository.Persist(ctx, &wagerTx, input.MessageID, input.ConsumerName)
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	money, err = domain.NewMoneyFromString(result.BalanceAmount, result.BalanceCurrency)
	if err != nil {
		return CreateWagerTransactionOutput{}, err
	}
	return CreateWagerTransactionOutput{
		TransactionID:    result.TransactionID,
		Status:           string(result.Status),
		Balance:          money,
		IdempotentReplay: result.IdempotentReplay,
	}, nil
}

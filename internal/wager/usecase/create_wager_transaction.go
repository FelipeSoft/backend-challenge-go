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
	ProviderID                     string
	ExternalTransactionID          string
	PlayerID                       string
	WalletID                       string
	RoundID                        *string
	GameID                         *string
	Kind                           string
	MoneyAmount                    string
	MoneyCurrency                  string
	IdempotencyKey                 string
	ReferenceExternalTransactionId *string
	MessageID                      *string
	ConsumerName                   *string
	PayloadHash                    *string
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

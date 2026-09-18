package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	walletdomain "github.com/FelipeSoft/backend-challenge-go/internal/wallet/domain/wallet"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet/repository"
	"github.com/google/uuid"
)

type CreateWallet struct {
	walletRepository repository.WalletRepository
}

type CreateWalletInput struct {
	PlayerID               string
	InitialBalanceAmount   string
	InitialBalanceCurrency string
	Now                    time.Time
}

type CreateWalletOutput struct {
	WalletID        string
	PlayerID        string
	BalanceAmount   string
	BalanceCurrency string
	Version         int64
}

func NewCreateWallet(walletRepository repository.WalletRepository) *CreateWallet {
	return &CreateWallet{
		walletRepository: walletRepository,
	}
}

func (uc *CreateWallet) Execute(ctx context.Context, input CreateWalletInput) (CreateWalletOutput, error) {
	initialMoney, err := domain.NewMoneyFromString(input.InitialBalanceAmount, input.InitialBalanceCurrency)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	walletUUID, err := uuid.NewV7()
	if err != nil {
		return CreateWalletOutput{}, err
	}
	wallet, err := walletdomain.RehydrateWallet(
		walletUUID.String(),
		input.PlayerID,
		initialMoney,
		1,
		input.Now,
		input.Now,
	)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	var wagerTransactionId string
	var payloadHash string
	if initialMoney.Amount() > 0 {
		txUUID, err := uuid.NewV7()
		if err != nil {
			return CreateWalletOutput{}, err
		}
		wagerTransactionId = txUUID.String()
		payloadMap := map[string]string{
			"playerId": input.PlayerID,
			"amount":   initialMoney.String(),
			"currency": initialMoney.Currency(),
			"kind":     "OPENING",
		}
		canonicalJSON, err := json.Marshal(payloadMap)
		if err != nil {
			return CreateWalletOutput{}, err
		}
		hashBytes := sha256.Sum256(canonicalJSON)
		payloadHash = fmt.Sprintf("%x", hashBytes)
	}
	_, err = uc.walletRepository.CreateWallet(ctx, wallet, wagerTransactionId, payloadHash)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	return CreateWalletOutput{
		WalletID:        wallet.ID(),
		PlayerID:        wallet.PlayerID(),
		BalanceAmount:   wallet.Balance().String(),
		BalanceCurrency: wallet.Currency(),
		Version:         wallet.Version(),
	}, nil
}
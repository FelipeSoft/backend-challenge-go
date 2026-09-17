package usecase

import (
	"context"
	"time"

	domain "github.com/FelipeSoft/jungle-gaming/internal/wallet/domain/wallet"
	"github.com/FelipeSoft/jungle-gaming/internal/wallet/repository"
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
	Version         int
}

func NewCreateWallet(walletRepository repository.WalletRepository) *CreateWallet {
	return &CreateWallet{
		walletRepository: walletRepository,
	}
}

func (uc *CreateWallet) Execute(ctx context.Context, input CreateWalletInput) (CreateWalletOutput, error) {
	wallet, err := domain.NewWallet(
		input.PlayerID,
		input.InitialBalanceCurrency,
		input.Now,
	)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	walletId, err := uc.walletRepository.CreateWallet(ctx, wallet)
	if err != nil {
		return CreateWalletOutput{}, err
	}
	return CreateWalletOutput{
		WalletID:        walletId,
		PlayerID:        input.PlayerID,
		BalanceAmount:   input.InitialBalanceAmount,
		BalanceCurrency: input.InitialBalanceCurrency,
		Version:         1,
	}, nil
}

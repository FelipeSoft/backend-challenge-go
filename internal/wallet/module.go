package wallet

import (
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet/query"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet/repository"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet/usecase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(repository.NewPostgresWalletRepository),
	fx.Provide(usecase.NewCreateWallet),
	fx.Provide(usecase.NewReconcileWallet),
	fx.Provide(query.NewGetWallet),
	fx.Provide(query.NewGetWalletLedger),
	fx.Provide(NewHandler),
)

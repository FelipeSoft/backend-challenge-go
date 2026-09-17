package deliveryhttp

import (
	"github.com/FelipeSoft/jungle-gaming/internal/provider"
	"github.com/FelipeSoft/jungle-gaming/internal/wagering"
	"github.com/FelipeSoft/jungle-gaming/internal/wallet"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In
	WalletHandler   *wallet.Handler
	WageringHandler *wagering.Handler
	ProviderHandler *provider.Handler
}

func NewRouter(p RouterParams) *gin.Engine {
	r := gin.Default()

	r.GET(
		"/wallets/:walletId", 
		p.WalletHandler.GetWallet,
	)
	r.GET(
		"/wallets/:walletId/ledger", 
		p.WalletHandler.GetWalletLedger,
	)
	r.POST(
		"/wallets", 
		p.WalletHandler.CreateWallet,
	)
	r.POST(
		"/wallets/:walletId/reconciliation", 
		p.WalletHandler.ReconcileWallet,
	)

	r.GET(
		"/wagering/transactions/:transactionId", 
		p.WageringHandler.GetTransaction,
	)
	r.POST(
		"/wagering/transactions", 
		p.WageringHandler.CreateWagerTransaction,
	)

	r.GET(
		"/providers/:providerId/wagering/transactions/:externalTransactionId", 
		p.ProviderHandler.GetExternalTransaction,
	)

	return r
}
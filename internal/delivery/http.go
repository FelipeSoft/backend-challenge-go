package delivery

import (
	"github.com/FelipeSoft/backend-challenge-go/internal/provider"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var HttpModule = fx.Options(
	fx.Provide(NewRouter),
)

type RouterParams struct {
	fx.In
	WalletHandler   *wallet.Handler
	WagerHandler    *wager.Handler
	ProviderHandler *provider.Handler
	AuthMiddleware  gin.HandlerFunc
}

func NewRouter(p RouterParams) *gin.Engine {
	r := gin.Default()
	r.GET("/health/live", func(c *gin.Context) {
		c.Status(200)
	})
	r.GET("/health/ready", func(c *gin.Context) {
		c.Status(200)
	})
	protected := r.Group("/")
	protected.Use(p.AuthMiddleware)
	{
		RegisterWalletRoutes(protected, p.WalletHandler)
		RegisterWageringRoutes(protected, p.WagerHandler)
		RegisterProviderRoutes(protected, p.ProviderHandler)
	}
	return r
}

func RegisterWalletRoutes(r *gin.RouterGroup, h *wallet.Handler) {
	r.POST("/wallets", h.CreateWallet)
	r.GET("/wallets/:walletId", h.GetWallet)
	r.GET("/wallets/:walletId/ledger", h.GetWalletLedger)
	r.POST("/wallets/:walletId/reconciliation", h.ReconcileWallet)
}

func RegisterWageringRoutes(r *gin.RouterGroup, h *wager.Handler) {
	r.POST("/wagering/transactions", h.CreateWagerTransaction)
	r.GET("/wagering/transactions/:transactionId", h.GetTransaction)
}

func RegisterProviderRoutes(r *gin.RouterGroup, h *provider.Handler) {
	r.GET("/providers/:providerId/wagering/transactions/:externalTransactionId", h.GetExternalTransaction)
}

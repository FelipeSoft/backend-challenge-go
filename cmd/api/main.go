package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/FelipeSoft/backend-challenge-go/internal/delivery"
	"github.com/FelipeSoft/backend-challenge-go/internal/platform"
	"github.com/FelipeSoft/backend-challenge-go/internal/provider"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet"
)

func main() {
	fx.New(
		platform.ConfigModule,
		platform.DatabaseModule,
		delivery.HttpModule,
		wallet.Module,
		wager.Module,
		provider.Module,
		fx.Invoke(func(lc fx.Lifecycle, r *gin.Engine) {
			port := os.Getenv("PORT")
			if port == "" {
				panic("missing port")
			}
			srv := &http.Server{
				Addr:    ":" + port,
				Handler: r,
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
							panic("error to start server: " + err.Error())
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
					defer cancel()
					return srv.Shutdown(shutdownCtx)
				},
			})
		}),
	).Run()
}
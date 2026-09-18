package wager

import (
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/query"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/repository"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/usecase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(repository.NewPostgresWagerRepository),
	fx.Provide(usecase.NewCreateWagerTransaction),
	fx.Provide(query.NewGetWagerTransaction),
	fx.Provide(NewHandler),
)

package provider

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewGetProviderWagerTransaction),
	fx.Provide(NewHandler),
)

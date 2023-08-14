package app

import (
	"sampleApp/gateway/cats"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"app gateways",
	fx.Option(
		fx.Provide(cats.New),
	),
)

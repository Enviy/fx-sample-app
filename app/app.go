package app

import (
	"fx-sample-app/gateway/redis"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"app gateways",
	fx.Option(
		fx.Provide(redis.New),
	),
)

package app

import (
	"fx-sample-app/gateway/cats"
	"fx-sample-app/gateway/redis"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"app gateways",
	fx.Provide(
		redis.New,
		cats.New,
	),
)

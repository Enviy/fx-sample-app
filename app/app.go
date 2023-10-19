package app

import (
	"fx-sample-app/gateway/cats"
	"fx-sample-app/gateway/redis"
	"fx-sample-app/repository/cache"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"app gateways",
	fx.Provide(
		redis.New,
		cache.New,
		cats.New,
	),
)

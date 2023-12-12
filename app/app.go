package app

import (
	"fx-sample-app/config"
	"fx-sample-app/gateway/cats"
	"fx-sample-app/gateway/redis"
	"fx-sample-app/gateway/slack"
	"fx-sample-app/logger"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"app base and gateways",
	fx.Provide(
		redis.New,
		cats.New,
		slack.New,
	),
	logger.Module,
	config.Module,
)

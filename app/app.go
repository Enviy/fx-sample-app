package app

import (
	"go.uber.org/fx"

	"fx-sample-app/config"
	"fx-sample-app/gateway/cats"
	"fx-sample-app/gateway/postgres"
	"fx-sample-app/gateway/redis"
	"fx-sample-app/gateway/slack"
	"fx-sample-app/logger"
)

var Module = fx.Module(
	"app base and gateways",
	fx.Provide(
		redis.New,
		cats.New,
		slack.New,
		postgres.New,
	),
	logger.Module,
	config.Module,
)

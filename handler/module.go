package handler

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"handler",
	fx.Option(
		fx.Invoke(New),
	),
)

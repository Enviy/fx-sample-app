package main

import (
	"sampleApp/app"
	"sampleApp/config"
	"sampleApp/controller"
	"sampleApp/handler"
	"sampleApp/infra"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,     // provide config.Provider.
		infra.Module,      // provide server, logger.
		app.Module,        // provide gateways.
		controller.Module, // provide controller interface.
		handler.Module,    // wire up to handlers.
	).Run()
}

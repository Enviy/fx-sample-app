package main

import (
	"fx-sample-app/app"
	"fx-sample-app/config"
	"fx-sample-app/controller"
	"fx-sample-app/handler"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,     // provide config.Provider.
		app.Module,        // provide gateways.
		controller.Module, // provide controller interface.
		handler.Module,    // wire up to handlers.
	).Run()
}

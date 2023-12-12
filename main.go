package main

import (
	"fx-sample-app/app"
	"fx-sample-app/controller"
	"fx-sample-app/handler"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		app.Module,        // provide gateways.
		controller.Module, // provide controller interface.
		handler.Module,    // wire up to handlers.
	).Run()
}

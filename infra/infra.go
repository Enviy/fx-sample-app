package infra

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module makes mux and logger available to the app.
var Module = fx.Module(
	"infra",
	fx.Provide(
		NewMux,
		NewLogger,
	),
)

func NewLogger() *zap.Logger {
	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.FunctionKey = "method"
	logger := zap.Must(logCfg.Build())
}

func NewMux(lc fx.Lifecycle, cfg config.Provider, log *zap.Logger) *http.ServeMux {
	mux := &http.ServeMux{}
	server := &http.Server{
		Addr:    cfg.Get("server.addr").String(),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			msg := "Started HTTP server."
			logger.Info(msg, zap.String("address", server.Addr))
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			go server.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server.")
			return server.Shutdown(ctx)
		},
	})

	return mux, logger
}

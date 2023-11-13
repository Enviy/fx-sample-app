package infra

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	rprof "runtime/pprof"
	"runtime/trace"

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

	return logger
}

func NewMux(lc fx.Lifecycle, cfg config.Provider, log *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	addr := cfg.Get("server.addr").String()
	port := cfg.Get("server.port").String()

	server := &http.Server{
		Addr:    addr + ":" + port,
		Handler: mux,
	}

	f, err := os.Create("profile.prof")
	if err != nil {
		log.Error("profile create", zap.Error(err))
		return nil
	}

	err = rprof.StartCPUProfile(f)
	if err != nil {
		log.Error("profile start", zap.Error(err))
		return nil
	}

	memprof, err := os.Create("mem.prof")
	if err != nil {
		log.Error("mem prof create", zap.Error(err))
		return nil
	}

	traceFile, err := os.Create("trace.out")
	if err != nil {
		log.Error("trace create", zap.Error(err))
		return nil
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			msg := "Started HTTP server."
			log.Info(msg, zap.String("address", server.Addr))

			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			go server.Serve(ln)
			// Goroutine for profiler.
			go func() {
				err = rprof.WriteHeapProfile(memprof)
				if err != nil {
					log.Error("write heap profile", zap.Error(err))
					return
				}

				err = trace.Start(traceFile)
				if err != nil {
					log.Error("trace start", zap.Error(err))
					return
				}
				defer trace.Stop()

				log.Info("pprof",
					zap.Any("output", http.ListenAndServe("localhost:6060", nil)),
				)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			f.Close()
			memprof.Close()
			traceFile.Close()
			rprof.StopCPUProfile()

			log.Info("Stopping HTTP server.")
			return server.Shutdown(ctx)
		},
	})

	return mux
}

package main

import (
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/IlianBuh/Post-service/internal/app"
	"github.com/IlianBuh/Post-service/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.New()

	log := setUpLogger(cfg.Env, os.Stdout)

	log.Info("logger was initialized", slog.Any("cfg", cfg))

	application := app.New(log, cfg.GRPC, cfg.Storage, cfg.UserProvider, cfg.Kafka, cfg.EventWorker)

	application.Start()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("signal recieved", slog.Any("signal", sign))

	application.Stop()
}

// setUpLogger returns set logger according to current environment
func setUpLogger(env string, w io.Writer) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

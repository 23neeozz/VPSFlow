package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/bosscloud/bosscloud/services/gateway/internal/adapter/http"
	"github.com/bosscloud/bosscloud/services/gateway/internal/config"
	"github.com/bosscloud/bosscloud/libs/go/observability"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	obs, err := observability.NewProvider(ctx, observability.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		LogLevel:       cfg.LogLevel,
		OTLPEndpoint:   cfg.OTLPEndpoint,
		EnableTracing:  cfg.OTLPEndpoint != "",
		EnableMetrics:  true,
	})
	if err != nil {
		slog.Error("failed to initialize observability", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := httpserver.NewServer(cfg, obs.Logger)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			obs.Logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case sig := <-sigCh:
		obs.Logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		obs.Logger.Error("http shutdown failed", slog.String("error", err.Error()))
	}

	if err := obs.Shutdown(shutdownCtx); err != nil {
		obs.Logger.Error("observability shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	obs.Logger.Info("gateway stopped gracefully")
}

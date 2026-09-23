package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/bosscloud/bosscloud/services/cluster/internal/adapter/http"
	"github.com/bosscloud/bosscloud/services/cluster/internal/adapter/repository/postgres"
	"github.com/bosscloud/bosscloud/services/cluster/internal/config"
	"github.com/bosscloud/bosscloud/services/cluster/internal/usecase"
	"github.com/bosscloud/bosscloud/libs/go/observability"
	"github.com/bosscloud/bosscloud/libs/go/security/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	obs, err := observability.NewProvider(ctx, observability.Config{
		ServiceName: cfg.ServiceName, ServiceVersion: cfg.ServiceVersion, Environment: cfg.Environment,
		LogLevel: cfg.LogLevel, OTLPEndpoint: cfg.OTLPEndpoint, EnableTracing: cfg.OTLPEndpoint != "", EnableMetrics: true,
	})
	if err != nil {
		slog.Error("failed to initialize observability", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := postgres.Up(cfg.DatabaseURL); err != nil {
		obs.Logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		obs.Logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtService, err := jwt.NewService(jwt.Config{SigningKey: []byte(cfg.JWTSigningKey), Issuer: cfg.JWTIssuer, Audience: cfg.JWTAudience})
	if err != nil {
		obs.Logger.Error("failed to initialize jwt service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	clusterService := usecase.NewService(postgres.NewStore(pool), cfg.HeartbeatStaleDur)
	server := httpserver.NewServer(cfg, obs.Logger, clusterService, jwtService)

	errCh := make(chan error, 1)
	go func() { errCh <- server.Listen() }()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		if err != nil {
			obs.Logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-sigCh:
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()
	_ = server.Shutdown()
	_ = obs.Shutdown(shutdownCtx)
}

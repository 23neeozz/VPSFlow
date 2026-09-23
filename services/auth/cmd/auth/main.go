package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/bosscloud/bosscloud/services/auth/internal/adapter/http"
	"github.com/bosscloud/bosscloud/services/auth/internal/adapter/repository/postgres"
	"github.com/bosscloud/bosscloud/services/auth/internal/config"
	"github.com/bosscloud/bosscloud/services/auth/internal/usecase"
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

	jwtService, err := jwt.NewService(jwt.Config{
		SigningKey: []byte(cfg.JWTSigningKey),
		Issuer:     cfg.JWTIssuer,
		Audience:   cfg.JWTAudience,
		AccessTTL:  cfg.AccessTokenTTL,
	})
	if err != nil {
		obs.Logger.Error("failed to initialize jwt service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	store := postgres.NewStore(pool)
	authService := usecase.NewService(
		store,
		store,
		store,
		store,
		store,
		jwtService,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.SessionTTL,
		cfg.MFAChallengeTTL,
		cfg.MaxLoginAttempts,
		cfg.LoginLockout,
	)

	server := httpserver.NewServer(cfg, obs.Logger, authService)

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

	obs.Logger.Info("auth service stopped gracefully")
}

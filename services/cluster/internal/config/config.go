package config

import (
	"fmt"
	"time"

	libconfig "github.com/bosscloud/bosscloud/libs/go/config"
)

type Config struct {
	HTTPAddr           string
	Environment        string
	LogLevel           string
	ServiceName        string
	ServiceVersion     string
	OTLPEndpoint       string
	DatabaseURL        string
	InternalAPIKey     string
	JWTSigningKey      string
	JWTIssuer          string
	JWTAudience        string
	HeartbeatStaleDur  time.Duration
	ShutdownTimeout    time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
}

func loadEnv() error {
	for _, file := range []string{"../../.env", ".env"} {
		if err := libconfig.NewLoader(file).Load(); err != nil {
			return fmt.Errorf("load environment: %w", err)
		}
	}
	return nil
}

func Load() (Config, error) {
	if err := loadEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:          libconfig.String("CLUSTER_HTTP_ADDR", ":8084"),
		Environment:       libconfig.String("CLUSTER_ENV", "development"),
		LogLevel:          libconfig.String("CLUSTER_LOG_LEVEL", "info"),
		ServiceName:       libconfig.String("CLUSTER_SERVICE_NAME", "cluster"),
		ServiceVersion:    libconfig.String("CLUSTER_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:      libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:       libconfig.String("CLUSTER_DATABASE_URL", "postgres://bosscloud:bosscloud_dev@localhost:5432/bosscloud_cluster?sslmode=disable"),
		InternalAPIKey:    libconfig.String("INTERNAL_API_KEY", "dev-internal-api-key"),
		JWTSigningKey:     libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:         libconfig.String("JWT_ISSUER", "bosscloud"),
		JWTAudience:       libconfig.String("JWT_AUDIENCE", "bosscloud-api"),
		HeartbeatStaleDur: libconfig.Duration("CLUSTER_HEARTBEAT_STALE", 90*time.Second),
		ShutdownTimeout:   libconfig.Duration("CLUSTER_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:       libconfig.Duration("CLUSTER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      libconfig.Duration("CLUSTER_WRITE_TIMEOUT", 10*time.Second),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("CLUSTER_DATABASE_URL must be set")
	}
	return cfg, nil
}

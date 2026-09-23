package config

import (
	"fmt"
	"time"

	libconfig "github.com/bosscloud/bosscloud/libs/go/config"
)

// Config holds iam service configuration.
type Config struct {
	HTTPAddr        string
	Environment     string
	LogLevel        string
	ServiceName     string
	ServiceVersion  string
	OTLPEndpoint    string
	DatabaseURL     string
	JWTSigningKey   string
	JWTIssuer       string
	JWTAudience     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func loadEnv() error {
	for _, file := range []string{"../../.env", ".env"} {
		if err := libconfig.NewLoader(file).Load(); err != nil {
			return fmt.Errorf("load environment: %w", err)
		}
	}
	return nil
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	if err := loadEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:        libconfig.String("IAM_HTTP_ADDR", ":8082"),
		Environment:     libconfig.String("IAM_ENV", "development"),
		LogLevel:        libconfig.String("IAM_LOG_LEVEL", "info"),
		ServiceName:     libconfig.String("IAM_SERVICE_NAME", "iam"),
		ServiceVersion:  libconfig.String("IAM_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:    libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:     libconfig.String("IAM_DATABASE_URL", "postgres://bosscloud:bosscloud_dev@localhost:5432/bosscloud_iam?sslmode=disable"),
		JWTSigningKey:   libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:       libconfig.String("JWT_ISSUER", "bosscloud"),
		JWTAudience:     libconfig.String("JWT_AUDIENCE", "bosscloud-api"),
		ShutdownTimeout: libconfig.Duration("IAM_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     libconfig.Duration("IAM_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    libconfig.Duration("IAM_WRITE_TIMEOUT", 10*time.Second),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("IAM_DATABASE_URL must be set")
	}

	return cfg, nil
}

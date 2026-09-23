package config

import (
	"fmt"
	"time"

	libconfig "github.com/vpsflow/vpsflow/libs/go/config"
)

type Config struct {
	HTTPAddr        string
	Environment     string
	LogLevel        string
	ServiceName     string
	ServiceVersion  string
	OTLPEndpoint    string
	DatabaseURL     string
	IAMServiceURL   string
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

func Load() (Config, error) {
	if err := loadEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:        libconfig.String("TENANT_HTTP_ADDR", ":8083"),
		Environment:     libconfig.String("TENANT_ENV", "development"),
		LogLevel:        libconfig.String("TENANT_LOG_LEVEL", "info"),
		ServiceName:     libconfig.String("TENANT_SERVICE_NAME", "tenant"),
		ServiceVersion:  libconfig.String("TENANT_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:    libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:     libconfig.String("TENANT_DATABASE_URL", "postgres://vpsflow:vpsflow_dev@localhost:5432/vpsflow_tenant?sslmode=disable"),
		IAMServiceURL:   libconfig.String("IAM_SERVICE_URL", "http://127.0.0.1:8082"),
		JWTSigningKey:   libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:       libconfig.String("JWT_ISSUER", "vpsflow"),
		JWTAudience:     libconfig.String("JWT_AUDIENCE", "vpsflow-api"),
		ShutdownTimeout: libconfig.Duration("TENANT_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     libconfig.Duration("TENANT_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    libconfig.Duration("TENANT_WRITE_TIMEOUT", 10*time.Second),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("TENANT_DATABASE_URL must be set")
	}
	return cfg, nil
}

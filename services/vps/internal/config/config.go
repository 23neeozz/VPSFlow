package config

import (
	"fmt"
	"time"

	libconfig "github.com/vpsflow/vpsflow/libs/go/config"
)

type Config struct {
	HTTPAddr         string
	Environment      string
	LogLevel         string
	ServiceName      string
	ServiceVersion   string
	OTLPEndpoint     string
	DatabaseURL      string
	VMServiceURL     string
	ConsoleServiceURL string
	TenantServiceURL string
	JWTSigningKey    string
	JWTIssuer        string
	JWTAudience      string
	ShutdownTimeout  time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
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
		HTTPAddr:          libconfig.String("VPS_HTTP_ADDR", ":8088"),
		Environment:       libconfig.String("VPS_ENV", "development"),
		LogLevel:          libconfig.String("VPS_LOG_LEVEL", "info"),
		ServiceName:       libconfig.String("VPS_SERVICE_NAME", "vps"),
		ServiceVersion:    libconfig.String("VPS_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:      libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:       libconfig.String("VPS_DATABASE_URL", "postgres://vpsflow:vpsflow_dev@localhost:5432/vpsflow_vps?sslmode=disable"),
		VMServiceURL:      libconfig.String("VM_SERVICE_URL", "http://127.0.0.1:8085"),
		ConsoleServiceURL: libconfig.String("CONSOLE_SERVICE_URL", "http://127.0.0.1:8087"),
		TenantServiceURL:  libconfig.String("TENANT_SERVICE_URL", "http://127.0.0.1:8083"),
		JWTSigningKey:     libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:         libconfig.String("JWT_ISSUER", "vpsflow"),
		JWTAudience:       libconfig.String("JWT_AUDIENCE", "vpsflow-api"),
		ShutdownTimeout:   libconfig.Duration("VPS_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:       libconfig.Duration("VPS_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:      libconfig.Duration("VPS_WRITE_TIMEOUT", 30*time.Second),
	}
	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("VPS_DATABASE_URL must be set")
	}
	return cfg, nil
}

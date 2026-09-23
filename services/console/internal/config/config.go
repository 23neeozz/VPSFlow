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
	VMServiceURL    string
	ConsolePublicURL string
	TokenTTL        time.Duration
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
		HTTPAddr:         libconfig.String("CONSOLE_HTTP_ADDR", ":8087"),
		Environment:      libconfig.String("CONSOLE_ENV", "development"),
		LogLevel:         libconfig.String("CONSOLE_LOG_LEVEL", "info"),
		ServiceName:      libconfig.String("CONSOLE_SERVICE_NAME", "console"),
		ServiceVersion:   libconfig.String("CONSOLE_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:     libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:      libconfig.String("CONSOLE_DATABASE_URL", "postgres://vpsflow:vpsflow_dev@localhost:5432/vpsflow_console?sslmode=disable"),
		VMServiceURL:     libconfig.String("VM_SERVICE_URL", "http://127.0.0.1:8085"),
		ConsolePublicURL: libconfig.String("CONSOLE_PUBLIC_URL", "http://127.0.0.1:8087"),
		TokenTTL:         libconfig.Duration("CONSOLE_TOKEN_TTL", 5*time.Minute),
		JWTSigningKey:    libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:        libconfig.String("JWT_ISSUER", "vpsflow"),
		JWTAudience:      libconfig.String("JWT_AUDIENCE", "vpsflow-api"),
		ShutdownTimeout:  libconfig.Duration("CONSOLE_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:      libconfig.Duration("CONSOLE_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:     libconfig.Duration("CONSOLE_WRITE_TIMEOUT", 10*time.Second),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("CONSOLE_DATABASE_URL must be set")
	}
	return cfg, nil
}

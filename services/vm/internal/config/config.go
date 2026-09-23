package config

import (
	"fmt"
	"time"

	libconfig "github.com/vpsflow/vpsflow/libs/go/config"
)

type Config struct {
	HTTPAddr          string
	Environment       string
	LogLevel          string
	ServiceName       string
	ServiceVersion    string
	OTLPEndpoint      string
	DatabaseURL       string
	ClusterURL        string
	AgentControlURL   string
	InternalAPIKey    string
	JWTSigningKey     string
	JWTIssuer         string
	JWTAudience       string
	ShutdownTimeout   time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
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
		HTTPAddr:        libconfig.String("VM_HTTP_ADDR", ":8085"),
		Environment:     libconfig.String("VM_ENV", "development"),
		LogLevel:        libconfig.String("VM_LOG_LEVEL", "info"),
		ServiceName:     libconfig.String("VM_SERVICE_NAME", "vm"),
		ServiceVersion:  libconfig.String("VM_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:    libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:     libconfig.String("VM_DATABASE_URL", "postgres://vpsflow:vpsflow_dev@localhost:5432/vpsflow_vm?sslmode=disable"),
		ClusterURL:      libconfig.String("CLUSTER_SERVICE_URL", "http://127.0.0.1:8084"),
		AgentControlURL: libconfig.String("AGENT_CONTROL_SERVICE_URL", "http://127.0.0.1:8086"),
		InternalAPIKey:  libconfig.String("INTERNAL_API_KEY", "dev-internal-api-key"),
		JWTSigningKey:   libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:       libconfig.String("JWT_ISSUER", "vpsflow"),
		JWTAudience:     libconfig.String("JWT_AUDIENCE", "vpsflow-api"),
		ShutdownTimeout: libconfig.Duration("VM_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     libconfig.Duration("VM_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    libconfig.Duration("VM_WRITE_TIMEOUT", 30*time.Second),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("VM_DATABASE_URL must be set")
	}
	return cfg, nil
}

package config

import (
	"fmt"
	"time"

	libconfig "github.com/bosscloud/bosscloud/libs/go/config"
)

type Config struct {
	HTTPAddr        string
	Environment     string
	LogLevel        string
	ServiceName     string
	ServiceVersion  string
	OTLPEndpoint    string
	DatabaseURL     string
	ClusterURL      string
	InternalAPIKey  string
	CommandTimeout  time.Duration
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
		HTTPAddr:        libconfig.String("AGENT_CONTROL_HTTP_ADDR", ":8086"),
		Environment:     libconfig.String("AGENT_CONTROL_ENV", "development"),
		LogLevel:        libconfig.String("AGENT_CONTROL_LOG_LEVEL", "info"),
		ServiceName:     libconfig.String("AGENT_CONTROL_SERVICE_NAME", "agent-control"),
		ServiceVersion:  libconfig.String("AGENT_CONTROL_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:    libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:     libconfig.String("AGENT_CONTROL_DATABASE_URL", "postgres://bosscloud:bosscloud_dev@localhost:5432/bosscloud_agent_control?sslmode=disable"),
		ClusterURL:      libconfig.String("CLUSTER_SERVICE_URL", "http://127.0.0.1:8084"),
		InternalAPIKey:  libconfig.String("INTERNAL_API_KEY", "dev-internal-api-key"),
		CommandTimeout:  libconfig.Duration("AGENT_CONTROL_COMMAND_TIMEOUT", 2*time.Minute),
		ShutdownTimeout: libconfig.Duration("AGENT_CONTROL_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     libconfig.Duration("AGENT_CONTROL_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    libconfig.Duration("AGENT_CONTROL_WRITE_TIMEOUT", 30*time.Second),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("AGENT_CONTROL_DATABASE_URL must be set")
	}
	return cfg, nil
}

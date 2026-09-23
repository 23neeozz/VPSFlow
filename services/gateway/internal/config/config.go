package config

import (
	"fmt"
	"time"

	libconfig "github.com/vpsflow/vpsflow/libs/go/config"
)

// Config holds gateway service configuration loaded from environment.
type Config struct {
	HTTPAddr       string
	Environment    string
	LogLevel       string
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
	ShutdownTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	AuthServiceURL   string
	IAMServiceURL    string
	TenantServiceURL string
	ClusterServiceURL string
	VmServiceURL     string
	ConsoleServiceURL string
	VpsServiceURL     string
	AgentControlServiceURL string
	JWTSigningKey  string
	JWTIssuer      string
	JWTAudience    string
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
		HTTPAddr:        libconfig.String("GATEWAY_HTTP_ADDR", ":8080"),
		Environment:     libconfig.String("GATEWAY_ENV", "development"),
		LogLevel:        libconfig.String("GATEWAY_LOG_LEVEL", "info"),
		ServiceName:     libconfig.String("GATEWAY_SERVICE_NAME", "gateway"),
		ServiceVersion:  libconfig.String("GATEWAY_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:    libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		ShutdownTimeout: libconfig.Duration("GATEWAY_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     libconfig.Duration("GATEWAY_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    libconfig.Duration("GATEWAY_WRITE_TIMEOUT", 10*time.Second),
		AuthServiceURL:   libconfig.String("AUTH_SERVICE_URL", "http://127.0.0.1:8081"),
		IAMServiceURL:    libconfig.String("IAM_SERVICE_URL", "http://127.0.0.1:8082"),
		TenantServiceURL: libconfig.String("TENANT_SERVICE_URL", "http://127.0.0.1:8083"),
		ClusterServiceURL: libconfig.String("CLUSTER_SERVICE_URL", "http://127.0.0.1:8084"),
		VmServiceURL:     libconfig.String("VM_SERVICE_URL", "http://127.0.0.1:8085"),
		ConsoleServiceURL: libconfig.String("CONSOLE_SERVICE_URL", "http://127.0.0.1:8087"),
		VpsServiceURL:     libconfig.String("VPS_SERVICE_URL", "http://127.0.0.1:8088"),
		AgentControlServiceURL: libconfig.String("AGENT_CONTROL_SERVICE_URL", "http://127.0.0.1:8086"),
		JWTSigningKey:   libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:       libconfig.String("JWT_ISSUER", "vpsflow"),
		JWTAudience:     libconfig.String("JWT_AUDIENCE", "vpsflow-api"),
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("GATEWAY_HTTP_ADDR must not be empty")
	}

	return cfg, nil
}

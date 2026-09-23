package config

import (
	"fmt"
	"time"

	libconfig "github.com/bosscloud/bosscloud/libs/go/config"
)

// Config holds auth service configuration.
type Config struct {
	HTTPAddr        string
	Environment     string
	LogLevel        string
	ServiceName     string
	ServiceVersion  string
	OTLPEndpoint    string
	DatabaseURL     string
	MigrationsPath  string
	JWTSigningKey   string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	SessionTTL      time.Duration
	MFAChallengeTTL time.Duration
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxLoginAttempts int
	LoginLockout     time.Duration
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
		HTTPAddr:         libconfig.String("AUTH_HTTP_ADDR", ":8081"),
		Environment:      libconfig.String("AUTH_ENV", "development"),
		LogLevel:         libconfig.String("AUTH_LOG_LEVEL", "info"),
		ServiceName:      libconfig.String("AUTH_SERVICE_NAME", "auth"),
		ServiceVersion:   libconfig.String("AUTH_SERVICE_VERSION", "0.1.0"),
		OTLPEndpoint:     libconfig.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		DatabaseURL:      libconfig.String("AUTH_DATABASE_URL", "postgres://bosscloud:bosscloud_dev@localhost:5432/bosscloud_auth?sslmode=disable"),
		MigrationsPath:   libconfig.String("AUTH_MIGRATIONS_PATH", "migrations"),
		JWTSigningKey:    libconfig.String("JWT_SIGNING_KEY", ""),
		JWTIssuer:        libconfig.String("JWT_ISSUER", "bosscloud"),
		JWTAudience:      libconfig.String("JWT_AUDIENCE", "bosscloud-api"),
		AccessTokenTTL:   libconfig.Duration("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:  libconfig.Duration("AUTH_REFRESH_TOKEN_TTL", 7*24*time.Hour),
		SessionTTL:       libconfig.Duration("AUTH_SESSION_TTL", 7*24*time.Hour),
		MFAChallengeTTL:  libconfig.Duration("AUTH_MFA_CHALLENGE_TTL", 5*time.Minute),
		ShutdownTimeout:  libconfig.Duration("AUTH_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:      libconfig.Duration("AUTH_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:     libconfig.Duration("AUTH_WRITE_TIMEOUT", 10*time.Second),
		MaxLoginAttempts: libconfig.Int("AUTH_MAX_LOGIN_ATTEMPTS", 5),
		LoginLockout:     libconfig.Duration("AUTH_LOGIN_LOCKOUT", 15*time.Minute),
	}

	if cfg.JWTSigningKey == "" {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be set")
	}
	if len(cfg.JWTSigningKey) < 32 {
		return Config{}, fmt.Errorf("JWT_SIGNING_KEY must be at least 32 characters")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("AUTH_DATABASE_URL must be set")
	}

	return cfg, nil
}

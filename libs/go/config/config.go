package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Loader loads configuration from environment variables with optional .env file support.
type Loader struct {
	envFile string
}

// NewLoader creates a configuration loader. Pass empty envFile to skip .env loading.
func NewLoader(envFile string) *Loader {
	return &Loader{envFile: envFile}
}

// Load attempts to load the .env file if configured. Missing file is not an error.
func (l *Loader) Load() error {
	if l.envFile == "" {
		return nil
	}
	if err := godotenv.Load(l.envFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load env file %q: %w", l.envFile, err)
	}
	return nil
}

// String returns the environment variable or defaultValue if unset.
func String(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// MustString returns the environment variable or panics if unset.
func MustString(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

// Int returns the environment variable as int or defaultValue if unset/invalid.
func Int(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// Bool returns the environment variable as bool or defaultValue if unset/invalid.
func Bool(key string, defaultValue bool) bool {
	v := strings.ToLower(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// Duration returns the environment variable as time.Duration or defaultValue.
func Duration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}

// Slice returns a comma-separated environment variable as a slice.
func Slice(key string, defaultValue []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

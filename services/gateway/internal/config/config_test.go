package config_test

import (
	"testing"

	"github.com/bosscloud/bosscloud/services/gateway/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GATEWAY_HTTP_ADDR", ":9090")
	t.Setenv("GATEWAY_SERVICE_NAME", "gateway-test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected :9090, got %s", cfg.HTTPAddr)
	}
	if cfg.ServiceName != "gateway-test" {
		t.Fatalf("expected gateway-test, got %s", cfg.ServiceName)
	}
}

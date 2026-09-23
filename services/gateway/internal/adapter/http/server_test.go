package http_test

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	httpserver "github.com/bosscloud/bosscloud/services/gateway/internal/adapter/http"
	"github.com/bosscloud/bosscloud/services/gateway/internal/config"
	"log/slog"
)

func TestHealthEndpoints(t *testing.T) {
	cfg := config.Config{
		ServiceName:    "gateway-test",
		ServiceVersion: "test",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httpserver.NewServer(cfg, log)

	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	req = httptest.NewRequest("GET", "/api/v1", nil)
	resp, err = server.App().Test(req)
	if err != nil {
		t.Fatalf("api request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

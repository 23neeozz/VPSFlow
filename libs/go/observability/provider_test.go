package observability

import (
	"context"
	"testing"
	"time"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()
	provider, err := NewProvider(ctx, Config{
		ServiceName:    "test-service",
		ServiceVersion: "0.0.1",
		Environment:    "test",
		LogLevel:       "debug",
		EnableTracing:  false,
		EnableMetrics:  true,
	})
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}
	if provider.Logger == nil {
		t.Fatal("expected logger")
	}
	if provider.MeterProvider == nil {
		t.Fatal("expected meter provider")
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

package config

import (
	"testing"
)

func TestString(t *testing.T) {
	t.Setenv("TEST_KEY", "value")
	if got := String("TEST_KEY", "default"); got != "value" {
		t.Fatalf("expected value, got %q", got)
	}
	if got := String("MISSING_KEY", "default"); got != "default" {
		t.Fatalf("expected default, got %q", got)
	}
}

func TestInt(t *testing.T) {
	t.Setenv("PORT", "8080")
	if got := Int("PORT", 3000); got != 8080 {
		t.Fatalf("expected 8080, got %d", got)
	}
	if got := Int("INVALID", 42); got != 42 {
		t.Fatalf("expected 42 for invalid, got %d", got)
	}
}

func TestBool(t *testing.T) {
	t.Setenv("ENABLED", "true")
	if got := Bool("ENABLED", false); !got {
		t.Fatal("expected true")
	}
}

func TestSlice(t *testing.T) {
	t.Setenv("HOSTS", "a,b, c")
	got := Slice("HOSTS", nil)
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("unexpected slice: %v", got)
	}
}

func TestLoaderLoadMissingFile(t *testing.T) {
	loader := NewLoader("nonexistent.env")
	if err := loader.Load(); err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
}

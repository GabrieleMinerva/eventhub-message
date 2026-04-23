package config

import (
	"os"
	"testing"
)

func TestLoadMockMode(t *testing.T) {
	t.Setenv("WORKER_MODE", "mock")
	t.Setenv("HTTP_PORT", "8181")
	t.Setenv("LOG_LEVEL", "debug")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != 8181 {
		t.Fatalf("expected 8181, got %d", cfg.HTTPPort)
	}
	if cfg.LogLevel.String() != "DEBUG" {
		t.Fatalf("expected DEBUG level, got %s", cfg.LogLevel.String())
	}
}

func TestLoadEventHubModeRequiresEnv(t *testing.T) {
	t.Setenv("WORKER_MODE", "eventhub")
	_ = os.Unsetenv("EVENTHUB_CONNECTION_STRING")
	_ = os.Unsetenv("EVENTHUB_NAME")
	_ = os.Unsetenv("EVENTHUB_EVENTS_FILE")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "trace")
	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid LOG_LEVEL error")
	}
}

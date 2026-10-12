package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCheckInterval(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "pkgctl")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"check_interval": "1h"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := Load()

	if err != nil {
		t.Fatal(err)
	}
	if settings.CheckInterval != time.Hour {
		t.Errorf("Expected 1h, got %v", settings.CheckInterval)
	}
}

func TestLoadWithoutFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	settings, err := Load()

	if err != nil {
		t.Fatal(err)
	}
	if settings.CheckInterval != 24*time.Hour {
		t.Errorf("Expected 24h, got %v", settings.CheckInterval)
	}
}

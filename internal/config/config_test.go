package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")

	cfg := &Config{VaultPath: "/test/vault"}
	if err := saveTo(cfg, path); err != nil {
		t.Fatalf("saveTo: %v", err)
	}

	got, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom: %v", err)
	}

	if got.VaultPath != cfg.VaultPath {
		t.Errorf("VaultPath = %q, want %q", got.VaultPath, cfg.VaultPath)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := loadFrom("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestSaveCreatesParentDirs(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "a", "b", "config.json")

	cfg := &Config{VaultPath: "/test/vault"}
	if err := saveTo(cfg, path); err != nil {
		t.Fatalf("saveTo: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

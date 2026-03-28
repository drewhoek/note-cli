package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	VaultPath string `json:"vault_path"`
}

// configPath returns the path to the config file using the OS-appropriate config directory.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "note", "config.json"), nil
}

// loadFrom reads and unmarshals a Config from the given file path.
func loadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("no config found — run: note config set vault-path <path>")
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// saveTo marshals cfg and writes it to the given file path, creating parent dirs as needed.
func saveTo(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads the config from the default OS config path.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	return loadFrom(path)
}

// Save writes cfg to the default OS config path.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return saveTo(cfg, path)
}

// Package config manages local TUI configuration stored in ~/.config/neto/config.json.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ErrNotConfigured is returned when the config file does not exist.
var ErrNotConfigured = errors.New("neto: config not found — run setup first")

// Config holds all TUI configuration values.
type Config struct {
	APIURL          string `json:"api_url"`
	SupabaseURL     string `json:"supabase_url"`
	SupabaseAnonKey string `json:"supabase_anon_key"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "neto", "config.json"), nil
}

// Load reads the config file from ~/.config/neto/config.json.
// Returns ErrNotConfigured if the file does not exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotConfigured
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to ~/.config/neto/config.json with permissions 0600.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

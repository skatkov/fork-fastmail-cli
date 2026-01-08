package tracking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds tracking configuration
type Config struct {
	Enabled          bool   `json:"enabled"`
	WorkerURL        string `json:"worker_url"`
	SecretsInKeyring bool   `json:"secrets_in_keyring,omitempty"`
	TrackingKey      string `json:"tracking_key,omitempty"`
	AdminKey         string `json:"admin_key,omitempty"`
}

// ConfigDir returns the shared email-tracking config directory
func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(configDir, "email-tracking"), nil
}

// ConfigPath returns the path to the tracking config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o700)
}

// LoadConfig loads tracking configuration from disk
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Enabled: false}, nil
		}
		return nil, fmt.Errorf("read tracking config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse tracking config: %w", err)
	}

	// Load secrets from keyring if configured
	if strings.TrimSpace(cfg.TrackingKey) == "" || strings.TrimSpace(cfg.AdminKey) == "" || cfg.SecretsInKeyring {
		trackingKey, adminKey, secretErr := LoadSecrets()
		if secretErr != nil {
			return nil, secretErr
		}

		if strings.TrimSpace(cfg.TrackingKey) == "" {
			cfg.TrackingKey = trackingKey
		}

		if strings.TrimSpace(cfg.AdminKey) == "" {
			cfg.AdminKey = adminKey
		}
	}

	return &cfg, nil
}

// SaveConfig saves tracking configuration to disk
func SaveConfig(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	toSave := cfg
	if cfg.SecretsInKeyring {
		s := *cfg
		s.TrackingKey = ""
		s.AdminKey = ""
		toSave = &s
	}

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tracking config: %w", err)
	}

	if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
		return fmt.Errorf("write tracking config: %w", writeErr)
	}

	return nil
}

// IsConfigured returns true if tracking is set up
func (c *Config) IsConfigured() bool {
	return c.Enabled && c.WorkerURL != "" && c.TrackingKey != ""
}

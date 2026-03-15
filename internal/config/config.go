// Package config handles loading and validation of git-hunk configuration.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "git-hunk", "config.toml"), nil
}

// Config holds the application configuration.
type Config struct {
	APIKey       string `toml:"api_key"`
	Model        string `toml:"model"`
	AutoGenerate bool   `toml:"auto_generate"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		Model:        "gpt-4o-mini",
		AutoGenerate: false,
	}
}

// Load reads the TOML config file at path and merges it with defaults.
// If the file does not exist, defaults are returned without error.
func Load(path string) (Config, error) {
	cfg := Default()
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		// Check for path error (file not found)
		var pe *os.PathError
		if errors.As(err, &pe) {
			return cfg, nil
		}
		return cfg, err
	}
	return cfg, nil
}

// LoadDefault loads configuration from the default path.
func LoadDefault() (Config, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return Default(), err
	}
	return Load(path)
}

// Validate returns an error if the configuration is invalid.
func (c Config) Validate() error {
	if c.Model == "" {
		return errors.New("model must not be empty")
	}
	return nil
}

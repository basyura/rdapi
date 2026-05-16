package config

import (
	"os"
	"path/filepath"
)

func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "rdapi", "config.toml")
	}
	return filepath.Join(home, ".config", "rdapi", "config.toml")
}

func GetDefaultSecretPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "secret.toml")
}

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func LoadAuth() (AuthConfig, error) {
	configPath := GetDefaultConfigPath()
	cfg, err := LoadAuthConfig(configPath)
	if err != nil {
		return AuthConfig{}, err
	}
	if err := LoadAuthSecrets(GetDefaultSecretPath(configPath), &cfg); err != nil {
		return AuthConfig{}, err
	}
	return cfg, nil
}

func LoadAuthConfig(path string) (AuthConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return AuthConfig{}, fmt.Errorf("read config: %w", err)
	}

	values := parseAuthSection(string(content))
	cfg := AuthConfig{
		ClientID:     values["client_id"],
		ClientSecret: values["client_secret"],
		RedirectURI:  values["redirect_uri"],
	}
	if cfg.ClientID == "" {
		return AuthConfig{}, errors.New("auth.client_id is missing in config")
	}
	if cfg.ClientSecret == "" {
		return AuthConfig{}, errors.New("auth.client_secret is missing in config")
	}
	return cfg, nil
}

func LoadAuthSecrets(path string, cfg *AuthConfig) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read secret: %w", err)
	}

	values := parseAuthSection(string(content))
	if values["access_token"] != "" {
		cfg.AccessToken = values["access_token"]
	}
	if values["refresh_token"] != "" {
		cfg.RefreshToken = values["refresh_token"]
	}
	return nil
}

func parseAuthSection(content string) map[string]string {
	values := make(map[string]string)
	inAuth := false

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inAuth = strings.TrimSpace(line[1:len(line)-1]) == "auth"
			continue
		}
		if !inAuth {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"`)
		values[key] = value
	}

	return values
}

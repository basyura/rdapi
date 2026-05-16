package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func LoadAuthSettings() (AuthSettings, error) {
	configPath := defaultConfigPath()
	cfg, err := loadAuthConfig(configPath)
	if err != nil {
		return AuthSettings{}, err
	}
	if err := loadAuthSecrets(defaultSecretPath(configPath), &cfg); err != nil {
		return AuthSettings{}, err
	}
	return cfg, nil
}

func loadAuthConfig(path string) (AuthSettings, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return AuthSettings{}, fmt.Errorf("read config: %w", err)
	}

	values := parseAuthSection(string(content))
	cfg := AuthSettings{
		ClientID:     values["client_id"],
		ClientSecret: values["client_secret"],
		RedirectURI:  values["redirect_uri"],
	}
	if cfg.ClientID == "" {
		return AuthSettings{}, errors.New("auth.client_id is missing in config")
	}
	if cfg.ClientSecret == "" {
		return AuthSettings{}, errors.New("auth.client_secret is missing in config")
	}
	return cfg, nil
}

func loadAuthSecrets(path string, cfg *AuthSettings) error {
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

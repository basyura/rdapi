package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SaveDefaultAuthTokens(accessToken, refreshToken string) error {
	configPath := GetDefaultConfigPath()
	return SaveAuthTokens(GetDefaultSecretPath(configPath), accessToken, refreshToken)
}

func SaveAuthTokens(path, accessToken, refreshToken string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read secret for token save: %w", err)
		}
		content = []byte("[auth]\n")
	}

	updated := UpsertAuthValue(string(content), "access_token", accessToken)
	if refreshToken != "" {
		updated = UpsertAuthValue(updated, "refresh_token", refreshToken)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create secret directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0600); err != nil {
		return fmt.Errorf("write secret tokens: %w", err)
	}
	return nil
}

func UpsertAuthValue(content, key, value string) string {
	lines := strings.Split(content, "\n")
	inAuth := false
	authStart := -1
	insertAt := len(lines)
	keyPrefix := key + " "
	replacement := fmt.Sprintf("%s = %q", key, value)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inAuth {
				insertAt = i
				break
			}
			inAuth = strings.TrimSpace(trimmed[1:len(trimmed)-1]) == "auth"
			if inAuth {
				authStart = i
				insertAt = i + 1
			}
			continue
		}
		if !inAuth {
			continue
		}
		insertAt = i + 1
		if strings.HasPrefix(trimmed, keyPrefix) || strings.HasPrefix(trimmed, key+"=") {
			lines[i] = replacement
			return strings.Join(lines, "\n")
		}
	}

	if authStart == -1 {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "[auth]", replacement)
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = replacement
	return strings.Join(lines, "\n")
}

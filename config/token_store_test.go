package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpsertAuthValue(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	secretPath := filepath.Join(dir, "secret.toml")
	content := `[auth]
client_id = "id"
client_secret = "secret"
redirect_uri = "http://localhost/callback"
`

	content = upsertAuthValue(content, "access_token", "access")
	content = upsertAuthValue(content, "refresh_token", "refresh")
	content = upsertAuthValue(content, "access_token", "updated")

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadAuthConfig(configPath)
	if err != nil {
		t.Fatalf("loadAuthConfig: %v", err)
	}
	if cfg.ClientID != "id" {
		t.Fatalf("client_id = %q", cfg.ClientID)
	}

	if err := os.WriteFile(secretPath, []byte(content), 0600); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	if err := loadAuthSecrets(secretPath, &cfg); err != nil {
		t.Fatalf("loadAuthSecrets: %v", err)
	}
	if cfg.AccessToken != "updated" {
		t.Fatalf("access_token = %q", cfg.AccessToken)
	}
	if cfg.RefreshToken != "refresh" {
		t.Fatalf("refresh_token = %q", cfg.RefreshToken)
	}
}

func TestSaveAndLoadAuthTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.toml")

	if err := saveAuthTokensToFile(path, "access", "refresh"); err != nil {
		t.Fatalf("saveAuthTokens: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat secret: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("mode = %v, want 0600", got)
	}

	var cfg AuthSettings
	if err := loadAuthSecrets(path, &cfg); err != nil {
		t.Fatalf("loadAuthSecrets: %v", err)
	}
	if cfg.AccessToken != "access" {
		t.Fatalf("access_token = %q", cfg.AccessToken)
	}
	if cfg.RefreshToken != "refresh" {
		t.Fatalf("refresh_token = %q", cfg.RefreshToken)
	}
}

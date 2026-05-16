package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"rdapi/api"
	"rdapi/config"
	"rdapi/term"
)

func TestExtractAuthorizationCode(t *testing.T) {
	tests := map[string]string{
		"abc":       "abc",
		"code=abc":  "abc",
		"?code=abc": "abc",
		"http://localhost/callback?code=abc&x=123": "abc",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got := api.ExtractAuthorizationCode(input)
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestUpsertAuthValue(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	secretPath := filepath.Join(dir, "secret.toml")
	content := `[auth]
client_id = "id"
client_secret = "secret"
redirect_uri = "http://localhost/callback"
`

	content = config.UpsertAuthValue(content, "access_token", "access")
	content = config.UpsertAuthValue(content, "refresh_token", "refresh")
	content = config.UpsertAuthValue(content, "access_token", "updated")

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := config.LoadAuthConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAuthConfig: %v", err)
	}
	if cfg.ClientID != "id" {
		t.Fatalf("client_id = %q", cfg.ClientID)
	}

	if err := os.WriteFile(secretPath, []byte(content), 0600); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	if err := config.LoadAuthSecrets(secretPath, &cfg); err != nil {
		t.Fatalf("LoadAuthSecrets: %v", err)
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

	if err := config.SaveAuthTokens(path, "access", "refresh"); err != nil {
		t.Fatalf("saveAuthTokens: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat secret: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("mode = %v, want 0600", got)
	}

	var cfg config.AuthConfig
	if err := config.LoadAuthSecrets(path, &cfg); err != nil {
		t.Fatalf("LoadAuthSecrets: %v", err)
	}
	if cfg.AccessToken != "access" {
		t.Fatalf("access_token = %q", cfg.AccessToken)
	}
	if cfg.RefreshToken != "refresh" {
		t.Fatalf("refresh_token = %q", cfg.RefreshToken)
	}
}

func TestGetDefaultSecretPath(t *testing.T) {
	got := config.GetDefaultSecretPath(filepath.Join("dir", "config.toml"))
	want := filepath.Join("dir", "secret.toml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTruncateByDisplayWidth(t *testing.T) {
	tests := []struct {
		name  string
		value string
		width int
		want  string
	}{
		{name: "fits", value: "hello", width: 5, want: "hello"},
		{name: "ascii", value: "hello world", width: 8, want: "hello …"},
		{name: "wide", value: "日本語タイトル", width: 7, want: "日本…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := term.TruncateByDisplayWidth(tt.value, tt.width)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if term.GetDisplayWidth(got) > tt.width {
				t.Fatalf("display width %d exceeds %d", term.GetDisplayWidth(got), tt.width)
			}
		})
	}
}

func TestFormatRaindropDate(t *testing.T) {
	item := api.Raindrop{Created: "2026-05-15T12:34:56.789Z"}

	got := formatRaindropDate(item)
	if got != "2026/05/15" {
		t.Fatalf("got %q, want %q", got, "2026/05/15")
	}
}

func TestRaindropCreatedAt(t *testing.T) {
	item := api.Raindrop{Created: "2026-05-15T12:34:56Z"}

	got := raindropCreatedAt(item)
	want := time.Date(2026, 5, 15, 12, 34, 56, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

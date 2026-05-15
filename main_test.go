package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
			got := extractAuthorizationCode(input)
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestUpsertAuthValue(t *testing.T) {
	content := `[auth]
client_id = "id"
client_secret = "secret"
redirect_uri = "http://localhost/callback"
`

	content = upsertAuthValue(content, "access_token", "access")
	content = upsertAuthValue(content, "refresh_token", "refresh")
	content = upsertAuthValue(content, "access_token", "updated")

	values := parseAuthSection(content)
	if values["access_token"] != "updated" {
		t.Fatalf("access_token = %q", values["access_token"])
	}
	if values["refresh_token"] != "refresh" {
		t.Fatalf("refresh_token = %q", values["refresh_token"])
	}
	if values["client_id"] != "id" {
		t.Fatalf("client_id = %q", values["client_id"])
	}
}

func TestSaveAndLoadAuthTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.toml")

	token := tokenResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	if err := saveAuthTokens(path, token); err != nil {
		t.Fatalf("saveAuthTokens: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat secret: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("mode = %v, want 0600", got)
	}

	var cfg authConfig
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

func TestDefaultSecretPath(t *testing.T) {
	got := defaultSecretPath(filepath.Join("dir", "config.toml"))
	want := filepath.Join("dir", "secret.toml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTruncateDisplayWidth(t *testing.T) {
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
			got := truncateDisplayWidth(tt.value, tt.width)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if displayWidth(got) > tt.width {
				t.Fatalf("display width %d exceeds %d", displayWidth(got), tt.width)
			}
		})
	}
}

func TestFormatRaindropDate(t *testing.T) {
	item := raindrop{Created: "2026-05-15T12:34:56.789Z"}

	got := formatRaindropDate(item)
	if got != "2026/05/15" {
		t.Fatalf("got %q, want %q", got, "2026/05/15")
	}
}

func TestRaindropCreatedAt(t *testing.T) {
	item := raindrop{Created: "2026-05-15T12:34:56Z"}

	got := raindropCreatedAt(item)
	want := time.Date(2026, 5, 15, 12, 34, 56, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

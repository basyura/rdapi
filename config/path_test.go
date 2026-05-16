package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultSecretPath(t *testing.T) {
	got := defaultSecretPath(filepath.Join("dir", "config.toml"))
	want := filepath.Join("dir", "secret.toml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

package main

import "testing"

func TestParseCommandLineFrom(t *testing.T) {
	got, err := parseCommandLine([]string{"--from", "20260515"})
	if err != nil {
		t.Fatalf("parseCommandLine() error = %v", err)
	}
	if got.fromSearch != "created:>2026-05-14" {
		t.Fatalf("fromSearch = %q, want created:>2026-05-14", got.fromSearch)
	}
}

func TestParseCommandLineRejectsInvalidFrom(t *testing.T) {
	_, err := parseCommandLine([]string{"--from", "2026-05-15"})
	if err == nil {
		t.Fatal("parseCommandLine() error = nil, want error")
	}
}

func TestParseCommandLineRejectsExtraArgument(t *testing.T) {
	_, err := parseCommandLine([]string{"--from", "20260515", "extra"})
	if err == nil {
		t.Fatal("parseCommandLine() error = nil, want error")
	}
}

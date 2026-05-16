package view

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"rdapi/api"
)

func TestFormatRaindropDate(t *testing.T) {
	item := api.Raindrop{
		CreatedAt: time.Date(2026, 5, 15, 12, 34, 56, 0, time.UTC),
	}

	got := formatRaindropDate(item)
	if got != "2026/05/15" {
		t.Fatalf("got %q, want %q", got, "2026/05/15")
	}
}

func TestFormatRaindropsDoesNotReorderInput(t *testing.T) {
	items := []api.Raindrop{
		{Title: "old", Link: "https://example.com/old", CreatedAt: time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)},
		{Title: "new", Link: "https://example.com/new", CreatedAt: time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)},
	}

	lines := FormatRaindrops(items, 80)

	if items[0].Title != "old" {
		t.Fatalf("items[0].Title = %q", items[0].Title)
	}
	if !strings.Contains(lines[0], "new") {
		t.Fatalf("lines = %v", lines)
	}
}

func TestFormatRaindropsIncludesURLLine(t *testing.T) {
	items := []api.Raindrop{
		{
			Title:     "B'z Official Website",
			Link:      "http://biz.com",
			CreatedAt: time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC),
		},
	}

	got := FormatRaindrops(items, 80)
	want := []string{
		"2026/01/09 : B'z Official Website",
		"             http://biz.com",
	}

	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestPrintLines(t *testing.T) {
	var out bytes.Buffer

	PrintLines(&out, []string{"one", "two"})

	if got := out.String(); got != "one\ntwo\n" {
		t.Fatalf("got %q", got)
	}
}

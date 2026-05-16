package api

import (
	"testing"
	"time"
)

func TestRaindropResponseRaindrop(t *testing.T) {
	item := raindropResponse{
		ID:      1,
		Title:   "title",
		Link:    "https://example.com",
		Domain:  "example.com",
		Created: "2026-05-15T12:34:56Z",
	}

	got := item.raindrop()
	wantCreatedAt := time.Date(2026, 5, 15, 12, 34, 56, 0, time.UTC)
	if got.ID != 1 || got.Title != "title" || got.Link != "https://example.com" || got.Domain != "example.com" {
		t.Fatalf("unexpected raindrop: %#v", got)
	}
	if !got.CreatedAt.Equal(wantCreatedAt) {
		t.Fatalf("CreatedAt = %s, want %s", got.CreatedAt, wantCreatedAt)
	}
}

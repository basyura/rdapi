package api

import (
	"io"
	"net/http"
	"strings"
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

func TestFetchRaindropsPageSetsSearchQuery(t *testing.T) {
	var gotRequest *http.Request
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotRequest = req
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"result":true,"items":[],"count":0}`)),
			}, nil
		}),
	}

	_, err := fetchRaindropsPage(client, "access-token", 2, "created:>2026-05-14")
	if err != nil {
		t.Fatalf("fetchRaindropsPage() error = %v", err)
	}

	if gotRequest == nil {
		t.Fatal("request was not sent")
	}
	query := gotRequest.URL.Query()
	if query.Get("page") != "2" {
		t.Fatalf("page = %q, want 2", query.Get("page"))
	}
	if query.Get("perpage") != "50" {
		t.Fatalf("perpage = %q, want 50", query.Get("perpage"))
	}
	if query.Get("search") != "created:>2026-05-14" {
		t.Fatalf("search = %q, want created:>2026-05-14", query.Get("search"))
	}
	if gotRequest.Header.Get("Authorization") != "Bearer access-token" {
		t.Fatalf("Authorization = %q", gotRequest.Header.Get("Authorization"))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

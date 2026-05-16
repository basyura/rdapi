package api

import "testing"

func TestExtractAuthorizationCode(t *testing.T) {
	tests := map[string]string{
		"abc":       "abc",
		"code=abc":  "abc",
		"?code=abc": "abc",
		"http://localhost/callback?code=abc&x=123": "abc",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got := ExtractAuthorizationCode(input)
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

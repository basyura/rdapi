package api

import (
	"net/url"
	"strings"
)

func ExtractAuthorizationCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if parsed, err := url.Parse(raw); err == nil {
		if code := parsed.Query().Get("code"); code != "" {
			return code
		}
	}
	if strings.HasPrefix(raw, "code=") || strings.Contains(raw, "&code=") {
		values, err := url.ParseQuery(strings.TrimPrefix(raw, "?"))
		if err == nil {
			if code := values.Get("code"); code != "" {
				return code
			}
		}
	}
	return raw
}

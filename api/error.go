package api

import (
	"fmt"
	"strings"
)

func responseError(operation, status string, body []byte) error {
	message := compactResponseBody(body)
	if message == "" {
		return fmt.Errorf("%s: %s", operation, status)
	}
	return fmt.Errorf("%s: %s: %s", operation, status, message)
}

func compactResponseBody(body []byte) string {
	message := strings.TrimSpace(string(body))
	if len(message) > 1024 {
		message = message[:1024]
	}
	return message
}

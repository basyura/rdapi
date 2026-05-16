package api

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const authorizeURL = "https://raindrop.io/oauth/authorize"

func CreateAuthorizationURL(clientID, redirectURI string) string {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	return authorizeURL + "?" + values.Encode()
}

func PromptAuthorizationCode(in io.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "Enter authorization code or redirected URL: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read authorization code: %w", err)
		}
		return "", errors.New("authorization code is required")
	}

	code := strings.TrimSpace(scanner.Text())
	if code == "" {
		return "", errors.New("authorization code is required")
	}
	return code, nil
}

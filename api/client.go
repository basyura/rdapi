package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"rdapi/config"
)

const (
	tokenURL     = "https://raindrop.io/oauth/access_token"
	raindropsURL = "https://api.raindrop.io/rest/v1/raindrops/0"
)

func ExchangeCode(client *http.Client, cfg config.AuthConfig, redirectURI, code string) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"redirect_uri":  redirectURI,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return TokenResponse{}, fmt.Errorf("encode token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("request access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenResponse{}, responseError("request access token", resp.Status, responseBody)
	}

	var token TokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return TokenResponse{}, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("access_token is missing in token response: %s", compactResponseBody(responseBody))
	}
	return token, nil
}

func RefreshAccessToken(client *http.Client, cfg config.AuthConfig) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": cfg.RefreshToken,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return TokenResponse{}, fmt.Errorf("encode refresh token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("refresh access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("read refresh token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenResponse{}, responseError("refresh access token", resp.Status, responseBody)
	}

	var token TokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return TokenResponse{}, fmt.Errorf("decode refresh token response: %w", err)
	}
	if token.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("access_token is missing in refresh token response: %s", compactResponseBody(responseBody))
	}
	if token.RefreshToken == "" {
		token.RefreshToken = cfg.RefreshToken
	}
	return token, nil
}

func FetchAllRaindrops(client *http.Client, accessToken string) ([]Raindrop, error) {
	var all []Raindrop
	for page := 0; ; page++ {
		items, err := FetchRaindropsPage(client, accessToken, page)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if len(items) < 50 {
			break
		}
	}
	return all, nil
}

func FetchRaindropsPage(client *http.Client, accessToken string, page int) ([]Raindrop, error) {
	endpoint, err := url.Parse(raindropsURL)
	if err != nil {
		return nil, fmt.Errorf("parse raindrops URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("page", strconv.Itoa(page))
	query.Set("perpage", "50")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create raindrops request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request raindrops: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read raindrops response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError("request raindrops", resp.Status, responseBody)
	}

	var result RaindropsResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("decode raindrops response: %w", err)
	}
	if !result.Result {
		return nil, errors.New("raindrops response result is false")
	}
	return result.Items, nil
}

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

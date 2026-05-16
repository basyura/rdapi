package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const tokenURL = "https://raindrop.io/oauth/access_token"

type TokenSet struct {
	AccessToken  string
	RefreshToken string
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func ExchangeCode(client *http.Client, clientID, clientSecret, redirectURI, code string) (TokenSet, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"redirect_uri":  redirectURI,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return TokenSet{}, fmt.Errorf("encode token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return TokenSet{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("request access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenSet{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenSet{}, responseError("request access token", resp.Status, responseBody)
	}

	var token tokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return TokenSet{}, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return TokenSet{}, fmt.Errorf("access_token is missing in token response: %s", compactResponseBody(responseBody))
	}
	return token.tokenSet(), nil
}

func RefreshAccessToken(client *http.Client, clientID, clientSecret, refreshToken string) (TokenSet, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return TokenSet{}, fmt.Errorf("encode refresh token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return TokenSet{}, fmt.Errorf("create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("refresh access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenSet{}, fmt.Errorf("read refresh token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenSet{}, responseError("refresh access token", resp.Status, responseBody)
	}

	var token tokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return TokenSet{}, fmt.Errorf("decode refresh token response: %w", err)
	}
	if token.AccessToken == "" {
		return TokenSet{}, fmt.Errorf("access_token is missing in refresh token response: %s", compactResponseBody(responseBody))
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token.tokenSet(), nil
}

func (t tokenResponse) tokenSet() TokenSet {
	return TokenSet{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}
}

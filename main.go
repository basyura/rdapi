package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	authorizeURL = "https://raindrop.io/oauth/authorize"
	tokenURL     = "https://raindrop.io/oauth/access_token"
	raindropsURL = "https://api.raindrop.io/rest/v1/raindrops/0"
)

type authConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AccessToken  string
	RefreshToken string
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type raindropsResponse struct {
	Result bool       `json:"result"`
	Items  []raindrop `json:"items"`
	Count  int        `json:"count"`
}

type raindrop struct {
	ID      int    `json:"_id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Domain  string `json:"domain"`
	Created string `json:"created"`
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, in io.Reader, out io.Writer) error {
	flags := flag.NewFlagSet("rdapi", flag.ContinueOnError)
	flags.SetOutput(out)
	configPath := flags.String("config", defaultConfigPath(), "config file path")
	secretPath := flags.String("secret", "", "secret file path")
	redirectURI := flags.String("redirect-uri", "", "OAuth redirect URI override")
	code := flags.String("code", "", "OAuth authorization code")
	if err := flags.Parse(args); err != nil {
		return err
	}

	cfg, err := loadAuthConfig(*configPath)
	if err != nil {
		return err
	}
	resolvedSecretPath := *secretPath
	if resolvedSecretPath == "" {
		resolvedSecretPath = defaultSecretPath(*configPath)
	}
	if err := loadAuthSecrets(resolvedSecretPath, &cfg); err != nil {
		return err
	}
	selectedRedirectURI := cfg.RedirectURI
	if *redirectURI != "" {
		selectedRedirectURI = *redirectURI
	}
	if selectedRedirectURI == "" {
		return errors.New("redirect_uri is required; add auth.redirect_uri to config or pass -redirect-uri with the same value used to obtain the code")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	accessToken := cfg.AccessToken
	forceAuthorization := *redirectURI != "" && *code == ""
	if accessToken == "" && cfg.RefreshToken != "" && !forceAuthorization {
		token, err := refreshAccessToken(client, cfg)
		if err != nil {
			return err
		}
		if err := saveAuthTokens(resolvedSecretPath, token); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}
	if *code == "" {
		if accessToken != "" && !forceAuthorization {
			return printRaindrops(client, accessToken, out)
		}
		authURL := authorizationURL(cfg.ClientID, selectedRedirectURI)
		if err := openBrowser(authURL); err != nil {
			return err
		}
		fmt.Fprintln(out, "Opened the authorization URL in your browser:")
		fmt.Fprintln(out, authURL)
		enteredCode, err := promptAuthorizationCode(in, out)
		if err != nil {
			return err
		}
		code = &enteredCode
	}

	authorizationCode := extractAuthorizationCode(*code)
	token, err := exchangeCode(client, cfg, selectedRedirectURI, authorizationCode)
	if err != nil {
		return err
	}
	if err := saveAuthTokens(resolvedSecretPath, token); err != nil {
		return err
	}

	return printRaindrops(client, token.AccessToken, out)
}

func promptAuthorizationCode(in io.Reader, out io.Writer) (string, error) {
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

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "rdapi", "config.toml")
	}
	return filepath.Join(home, ".config", "rdapi", "config.toml")
}

func defaultSecretPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "secret.toml")
}

func loadAuthConfig(path string) (authConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return authConfig{}, fmt.Errorf("read config: %w", err)
	}

	values := parseAuthSection(string(content))
	cfg := authConfig{
		ClientID:     values["client_id"],
		ClientSecret: values["client_secret"],
		RedirectURI:  values["redirect_uri"],
	}
	if cfg.ClientID == "" {
		return authConfig{}, errors.New("auth.client_id is missing in config")
	}
	if cfg.ClientSecret == "" {
		return authConfig{}, errors.New("auth.client_secret is missing in config")
	}
	return cfg, nil
}

func loadAuthSecrets(path string, cfg *authConfig) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read secret: %w", err)
	}

	values := parseAuthSection(string(content))
	if values["access_token"] != "" {
		cfg.AccessToken = values["access_token"]
	}
	if values["refresh_token"] != "" {
		cfg.RefreshToken = values["refresh_token"]
	}
	return nil
}

func parseAuthSection(content string) map[string]string {
	values := make(map[string]string)
	inAuth := false

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inAuth = strings.TrimSpace(line[1:len(line)-1]) == "auth"
			continue
		}
		if !inAuth {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"`)
		values[key] = value
	}

	return values
}

func authorizationURL(clientID, redirectURI string) string {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	return authorizeURL + "?" + values.Encode()
}

func openBrowser(rawURL string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		command = "xdg-open"
	}
	args = append(args, rawURL)

	if err := exec.Command(command, args...).Start(); err != nil {
		return fmt.Errorf("open authorization URL in browser: %w", err)
	}
	return nil
}

func extractAuthorizationCode(raw string) string {
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

func exchangeCode(client *http.Client, cfg authConfig, redirectURI, code string) (tokenResponse, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"redirect_uri":  redirectURI,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return tokenResponse{}, fmt.Errorf("encode token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("request access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return tokenResponse{}, responseError("request access token", resp.Status, responseBody)
	}

	var token tokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return tokenResponse{}, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return tokenResponse{}, fmt.Errorf("access_token is missing in token response: %s", compactResponseBody(responseBody))
	}
	return token, nil
}

func refreshAccessToken(client *http.Client, cfg authConfig) (tokenResponse, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": cfg.RefreshToken,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return tokenResponse{}, fmt.Errorf("encode refresh token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, &buf)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("refresh access token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("read refresh token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return tokenResponse{}, responseError("refresh access token", resp.Status, responseBody)
	}

	var token tokenResponse
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return tokenResponse{}, fmt.Errorf("decode refresh token response: %w", err)
	}
	if token.AccessToken == "" {
		return tokenResponse{}, fmt.Errorf("access_token is missing in refresh token response: %s", compactResponseBody(responseBody))
	}
	if token.RefreshToken == "" {
		token.RefreshToken = cfg.RefreshToken
	}
	return token, nil
}

func printRaindrops(client *http.Client, accessToken string, out io.Writer) error {
	items, err := fetchAllRaindrops(client, accessToken)
	if err != nil {
		return err
	}

	sort.SliceStable(items, func(i, j int) bool {
		return raindropCreatedAt(items[i]).After(raindropCreatedAt(items[j]))
	})

	width := terminalWidth()
	for _, item := range items {
		line := fmt.Sprintf("%s : %s", formatRaindropDate(item), item.Title)
		fmt.Fprintln(out, truncateDisplayWidth(line, width))
	}

	return nil
}

func formatRaindropDate(item raindrop) string {
	createdAt := raindropCreatedAt(item)
	if createdAt.IsZero() {
		return "0000/00/00"
	}
	return createdAt.Format("2006/01/02")
}

func raindropCreatedAt(item raindrop) time.Time {
	if item.Created == "" {
		return time.Time{}
	}
	createdAt, err := time.Parse(time.RFC3339, item.Created)
	if err == nil {
		return createdAt
	}
	createdAt, err = time.Parse("2006-01-02T15:04:05.000Z", item.Created)
	if err == nil {
		return createdAt
	}
	return time.Time{}
}

func fetchAllRaindrops(client *http.Client, accessToken string) ([]raindrop, error) {
	var all []raindrop
	for page := 0; ; page++ {
		items, err := fetchRaindropsPage(client, accessToken, page)
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

func fetchRaindropsPage(client *http.Client, accessToken string, page int) ([]raindrop, error) {
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

	var result raindropsResponse
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

func saveAuthTokens(path string, token tokenResponse) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read secret for token save: %w", err)
		}
		content = []byte("[auth]\n")
	}

	updated := upsertAuthValue(string(content), "access_token", token.AccessToken)
	if token.RefreshToken != "" {
		updated = upsertAuthValue(updated, "refresh_token", token.RefreshToken)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create secret directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0600); err != nil {
		return fmt.Errorf("write secret tokens: %w", err)
	}
	return nil
}

func upsertAuthValue(content, key, value string) string {
	lines := strings.Split(content, "\n")
	inAuth := false
	authStart := -1
	insertAt := len(lines)
	keyPrefix := key + " "
	replacement := fmt.Sprintf("%s = %q", key, value)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inAuth {
				insertAt = i
				break
			}
			inAuth = strings.TrimSpace(trimmed[1:len(trimmed)-1]) == "auth"
			if inAuth {
				authStart = i
				insertAt = i + 1
			}
			continue
		}
		if !inAuth {
			continue
		}
		insertAt = i + 1
		if strings.HasPrefix(trimmed, keyPrefix) || strings.HasPrefix(trimmed, key+"=") {
			lines[i] = replacement
			return strings.Join(lines, "\n")
		}
	}

	if authStart == -1 {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "[auth]", replacement)
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = replacement
	return strings.Join(lines, "\n")
}

func terminalWidth() int {
	output, err := exec.Command("stty", "size").Output()
	if err != nil {
		return 80
	}

	fields := strings.Fields(string(output))
	if len(fields) != 2 {
		return 80
	}

	width, err := strconv.Atoi(fields[1])
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func truncateDisplayWidth(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if displayWidth(value) <= maxWidth {
		return value
	}
	if maxWidth <= 1 {
		return "…"
	}

	ellipsis := "…"
	limit := maxWidth - displayWidth(ellipsis)
	var builder strings.Builder
	current := 0
	for _, r := range value {
		width := runeDisplayWidth(r)
		if current+width > limit {
			break
		}
		builder.WriteRune(r)
		current += width
	}
	builder.WriteString(ellipsis)
	return builder.String()
}

func displayWidth(value string) int {
	width := 0
	for _, r := range value {
		width += runeDisplayWidth(r)
	}
	return width
}

func runeDisplayWidth(r rune) int {
	if r == '\t' {
		return 4
	}
	if r < 0x20 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if utf8.RuneLen(r) > 1 {
		return 2
	}
	return 1
}

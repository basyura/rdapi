package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"rdapi/api"
	"rdapi/config"
	"rdapi/term"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer) error {
	cfg, err := config.LoadAuth()
	if err != nil {
		return err
	}
	if cfg.RedirectURI == "" {
		return errors.New("redirect_uri is required; add auth.redirect_uri to config")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	accessToken := cfg.AccessToken
	if accessToken == "" && cfg.RefreshToken != "" {
		token, err := api.RefreshAccessToken(client, cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken)
		if err != nil {
			return err
		}
		if err := config.SaveDefaultAuthTokens(token.AccessToken, token.RefreshToken); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}
	if accessToken == "" {
		authURL := api.CreateAuthorizationURL(cfg.ClientID, cfg.RedirectURI)
		if err := term.OpenBrowser(authURL); err != nil {
			return err
		}
		fmt.Fprintln(out, "Opened the authorization URL in your browser:")
		fmt.Fprintln(out, authURL)
		enteredCode, err := promptAuthorizationCode(in, out)
		if err != nil {
			return err
		}

		authorizationCode := api.ExtractAuthorizationCode(enteredCode)
		token, err := api.ExchangeCode(client, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, authorizationCode)
		if err != nil {
			return err
		}
		if err := config.SaveDefaultAuthTokens(token.AccessToken, token.RefreshToken); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}

	return printRaindrops(client, accessToken, out)
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

func printRaindrops(client *http.Client, accessToken string, out io.Writer) error {
	items, err := api.FetchAllRaindrops(client, accessToken)
	if err != nil {
		return err
	}

	sort.SliceStable(items, func(i, j int) bool {
		return raindropCreatedAt(items[i]).After(raindropCreatedAt(items[j]))
	})

	width := term.GetTerminalWidth()
	for _, item := range items {
		line := fmt.Sprintf("%s : %s", formatRaindropDate(item), item.Title)
		fmt.Fprintln(out, term.TruncateByDisplayWidth(line, width))
	}

	return nil
}

func formatRaindropDate(item api.Raindrop) string {
	createdAt := raindropCreatedAt(item)
	if createdAt.IsZero() {
		return "0000/00/00"
	}
	return createdAt.Format("2006/01/02")
}

func raindropCreatedAt(item api.Raindrop) time.Time {
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

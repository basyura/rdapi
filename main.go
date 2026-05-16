package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"rdapi/api"
	"rdapi/cli"
	"rdapi/config"
	"rdapi/term"
	"rdapi/view"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer) error {
	cfg, err := config.LoadAuthSettings()
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
		if err := config.SaveAuthTokens(token.AccessToken, token.RefreshToken); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}
	if accessToken == "" {
		authURL := api.BuildAuthorizationURL(cfg.ClientID, cfg.RedirectURI)
		if err := cli.OpenURL(authURL); err != nil {
			return err
		}
		fmt.Fprintln(out, "Opened the authorization URL in your browser:")
		fmt.Fprintln(out, authURL)
		enteredCode, err := cli.PromptAuthorizationCode(in, out)
		if err != nil {
			return err
		}

		authorizationCode := api.ExtractAuthorizationCode(enteredCode)
		token, err := api.ExchangeCode(client, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, authorizationCode)
		if err != nil {
			return err
		}
		if err := config.SaveAuthTokens(token.AccessToken, token.RefreshToken); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}

	return fetchAndPrintRaindrops(client, accessToken, out)
}

func fetchAndPrintRaindrops(client *http.Client, accessToken string, out io.Writer) error {
	items, err := api.FetchAllRaindrops(client, accessToken)
	if err != nil {
		return err
	}

	lines := view.FormatRaindrops(items, term.TerminalWidth())
	view.PrintLines(out, lines)
	return nil
}

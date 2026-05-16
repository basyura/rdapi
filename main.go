package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"rdapi/api"
	"rdapi/config"
	"rdapi/term"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, in io.Reader, out io.Writer) error {
	flags := flag.NewFlagSet("rdapi", flag.ContinueOnError)
	flags.SetOutput(out)
	configPath := flags.String("config", config.GetDefaultConfigPath(), "config file path")
	secretPath := flags.String("secret", "", "secret file path")
	redirectURI := flags.String("redirect-uri", "", "OAuth redirect URI override")
	code := flags.String("code", "", "OAuth authorization code")
	if err := flags.Parse(args); err != nil {
		return err
	}

	cfg, err := config.LoadAuthConfig(*configPath)
	if err != nil {
		return err
	}
	resolvedSecretPath := *secretPath
	if resolvedSecretPath == "" {
		resolvedSecretPath = config.GetDefaultSecretPath(*configPath)
	}
	if err := config.LoadAuthSecrets(resolvedSecretPath, &cfg); err != nil {
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
		token, err := api.RefreshAccessToken(client, cfg)
		if err != nil {
			return err
		}
		if err := api.SaveAuthTokens(resolvedSecretPath, token); err != nil {
			return err
		}
		accessToken = token.AccessToken
	}
	if *code == "" {
		if accessToken != "" && !forceAuthorization {
			return printRaindrops(client, accessToken, out)
		}
		authURL := api.CreateAuthorizationURL(cfg.ClientID, selectedRedirectURI)
		if err := api.OpenBrowser(authURL); err != nil {
			return err
		}
		fmt.Fprintln(out, "Opened the authorization URL in your browser:")
		fmt.Fprintln(out, authURL)
		enteredCode, err := api.PromptAuthorizationCode(in, out)
		if err != nil {
			return err
		}
		code = &enteredCode
	}

	authorizationCode := api.ExtractAuthorizationCode(*code)
	token, err := api.ExchangeCode(client, cfg, selectedRedirectURI, authorizationCode)
	if err != nil {
		return err
	}
	if err := api.SaveAuthTokens(resolvedSecretPath, token); err != nil {
		return err
	}

	return printRaindrops(client, token.AccessToken, out)
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

package main

import (
	"errors"
	"flag"
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
	if err := run(os.Stdin, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type commandLineOptions struct {
	fromSearch string
}

func run(in io.Reader, out io.Writer, args []string) error {
	options, err := parseCommandLine(args)
	if err != nil {
		return err
	}

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

	return fetchAndPrintRaindrops(client, accessToken, out, options.fromSearch)
}

func parseCommandLine(args []string) (commandLineOptions, error) {
	flags := flag.NewFlagSet("rdapi", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	from := flags.String("from", "", "fetch bookmarks created on or after yyyyMMdd")
	if err := flags.Parse(args); err != nil {
		return commandLineOptions{}, err
	}
	if flags.NArg() > 0 {
		return commandLineOptions{}, fmt.Errorf("unexpected argument: %s", flags.Arg(0))
	}
	if *from == "" {
		return commandLineOptions{}, nil
	}

	fromDate, err := parseFromDate(*from)
	if err != nil {
		return commandLineOptions{}, err
	}
	return commandLineOptions{fromSearch: "created:>" + fromDate}, nil
}

func parseFromDate(value string) (string, error) {
	if len(value) != len("20060102") {
		return "", fmt.Errorf("invalid --from date %q: use yyyyMMdd", value)
	}
	fromDate, err := time.Parse("20060102", value)
	if err != nil {
		return "", fmt.Errorf("invalid --from date %q: use yyyyMMdd", value)
	}
	return fromDate.AddDate(0, 0, -1).Format("2006-01-02"), nil
}

func fetchAndPrintRaindrops(client *http.Client, accessToken string, out io.Writer, search string) error {
	items, err := api.FetchAllRaindrops(client, accessToken, search)
	if err != nil {
		return err
	}

	lines := view.FormatRaindrops(items, term.TerminalWidth())
	view.PrintLines(out, lines)
	return nil
}

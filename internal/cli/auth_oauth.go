package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cli/browser"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/cli/internal/oauth"
	"github.com/spf13/cobra"
)

// loginViaOAuth runs the interactive OAuth browser flow and returns an
// access token. Production wiring constructs an oauth.Client from the
// active config and delegates to its Login method. Tests override this
// var directly to skip the listener / browser / token exchange.
var loginViaOAuth = defaultLoginViaOAuth

// isStdinTTY reports whether the command's stdin is a real terminal.
// Tests override it directly so they can drive the OAuth branch without
// faking a PTY. The underlying check delegates to isInteractiveInput
// (see confirm.go) so the OAuth branch stays in lockstep with how
// destructive-action prompts decide interactivity.
var isStdinTTY = func(cmd *cobra.Command) bool {
	return isInteractiveInput(cmd.InOrStdin())
}

// defaultLoginViaOAuth is the production implementation of the OAuth flow.
// It is wired through `loginViaOAuth` so the integration test in
// auth_oauth_test.go can swap it for a stub.
func defaultLoginViaOAuth(ctx context.Context, cfg *config.Config, errOut io.Writer) (string, error) {
	clientID := config.OAuthClientID(cfg.Sandbox)
	if clientID == "" {
		return "", oauth.ErrNotProvisioned
	}
	c := &oauth.Client{
		ClientID:      clientID,
		AuthorizeBase: config.AuthorizeURL(cfg.Sandbox),
		TokenURL:      config.OAuthTokenURL(cfg.BaseURL),
		BrowserOpener: browser.OpenURL,
		Stderr:        errOut,
	}
	return c.Login(ctx)
}

// acquireToken obtains the access token for a fresh `auth login`. With
// --with-token or non-TTY stdin it reads the token from stdin; on a TTY it
// runs the interactive OAuth browser flow. A browser-login failure is
// reported and returned, with no fall back to a token prompt: the error tells
// the user to retry or pass --with-token.
func acquireToken(cmd *cobra.Command, cfg *config.Config, withToken bool) (string, error) {
	switch {
	case withToken:
		return readLoginToken(cmd, true)
	case !isStdinTTY(cmd):
		return readLoginToken(cmd, false)
	}

	token, err := loginViaOAuth(context.Background(), cfg, cmd.ErrOrStderr())
	switch {
	case err == nil:
		return token, nil
	case errors.Is(err, context.Canceled):
		return "", err
	case errors.Is(err, oauth.ErrNotProvisioned):
		return "", errors.New("interactive browser login is not available in this build\n\nRun `dnsimple auth login --with-token` to authenticate with an API token instead")
	default:
		return "", fmt.Errorf("browser login failed: %w\n\nRetry `dnsimple auth login`, or run `dnsimple auth login --with-token` to authenticate with an API token instead", err)
	}
}

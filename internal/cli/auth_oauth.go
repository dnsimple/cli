package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cli/browser"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/cli/internal/oauth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// loginViaOAuth runs the interactive OAuth browser flow and returns an
// access token. Production wiring constructs an oauth.Client from the
// active config and delegates to its Login method. Tests override this
// var directly to skip the listener / browser / token exchange.
var loginViaOAuth = defaultLoginViaOAuth

// isStdinTTY reports whether the command's stdin is a real terminal. Tests
// override it directly so they can drive the OAuth branch without faking a
// PTY. The default mirrors the check used inside readLoginToken.
var isStdinTTY = func(cmd *cobra.Command) bool {
	f, ok := cmd.InOrStdin().(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
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
		TokenURL:      config.OAuthTokenURL(cfg.Sandbox),
		BrowserOpener: browser.OpenURL,
		Stderr:        errOut,
	}
	return c.Login(ctx)
}

// acquireToken obtains the access token for a fresh `auth login`. It
// dispatches between three paths:
//
//  1. --with-token: read the token as a single line from stdin (existing
//     path, used for piping a pre-issued token).
//  2. Stdin is not a TTY (CI, redirected input): read the token from
//     stdin, the same shape as --with-token but without requiring users
//     to remember the flag.
//  3. Stdin is a TTY: run the interactive OAuth browser flow.
//
// When the OAuth flow is not provisioned (empty client ID for the target
// environment in this build), the function falls back to the paste prompt
// with a one-line notice so users on builds shipped before the server-side
// rollout still get today's behaviour.
func acquireToken(cmd *cobra.Command, cfg *config.Config, withToken bool) (string, error) {
	switch {
	case withToken:
		return readLoginToken(cmd, true)
	case !isStdinTTY(cmd):
		return readLoginToken(cmd, false)
	}

	token, err := loginViaOAuth(context.Background(), cfg, cmd.ErrOrStderr())
	if errors.Is(err, oauth.ErrNotProvisioned) {
		fmt.Fprintln(cmd.ErrOrStderr(),
			"Interactive browser login is not yet available in this build. Falling back to API token paste.")
		return readLoginToken(cmd, false)
	}
	return token, err
}

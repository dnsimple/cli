package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/cli/internal/oauth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// stubLoginViaOAuth swaps the OAuth entry point for the duration of the
// test, restoring the original on Cleanup.
func stubLoginViaOAuth(t *testing.T, stub func(ctx context.Context, cfg *config.Config, errOut io.Writer) (string, error)) {
	t.Helper()
	prev := loginViaOAuth
	loginViaOAuth = stub
	t.Cleanup(func() { loginViaOAuth = prev })
}

// forceTTY makes isStdinTTY return the given value for the duration of the
// test. Tests that exercise the OAuth branch need to flip it to true
// because strings.NewReader is not an *os.File.
func forceTTY(t *testing.T, tty bool) {
	t.Helper()
	prev := isStdinTTY
	isStdinTTY = func(*cobra.Command) bool { return tty }
	t.Cleanup(func() { isStdinTTY = prev })
}

// --- acquireToken branching ---

func TestAcquireTokenWithTokenFlagReadsFromStdin(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("tok-from-stdin\n"))
	cmd.SetErr(io.Discard)

	got, err := acquireToken(cmd, &config.Config{}, true)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "tok-from-stdin", got)
}

func TestAcquireTokenNonTTYReadsFromStdin(t *testing.T) {
	forceTTY(t, false)
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("tok-piped\n"))
	cmd.SetErr(io.Discard)

	got, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "tok-piped", got)
}

func TestAcquireTokenTTYRunsOAuth(t *testing.T) {
	forceTTY(t, true)

	var capturedSandbox bool
	stubLoginViaOAuth(t, func(_ context.Context, cfg *config.Config, _ io.Writer) (string, error) {
		capturedSandbox = cfg.Sandbox
		return "tok-from-oauth", nil
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("")) // OAuth path must not consume stdin
	cmd.SetErr(io.Discard)

	got, err := acquireToken(cmd, &config.Config{Sandbox: true}, false)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "tok-from-oauth", got)
	assert.True(t, capturedSandbox, "OAuth flow should receive cfg.Sandbox=true")
}

func TestAcquireTokenErrorsOnErrNotProvisioned(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", oauth.ErrNotProvisioned
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("would-be-pasted\n")) // must not be consumed
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "not available in this build")
	assert.Contains(t, err.Error(), "--with-token")
}

func TestAcquireTokenErrorsOnTransientOAuthError(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", fmt.Errorf("network: connection refused")
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("would-be-pasted\n")) // must not be consumed
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "browser login failed")
	assert.Contains(t, err.Error(), "connection refused")
	assert.Contains(t, err.Error(), "--with-token")
}

// TestAcquireTokenAbortsOnAccessDenied pins that a user who explicitly
// denied consent in the browser is NOT pestered with a paste prompt.
func TestAcquireTokenAbortsOnAccessDenied(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", &oauth.AuthError{Code: "access_denied", Description: "user cancelled"}
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("would-be-pasted\n"))
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.Error(t, err) {
		return
	}
	var ae *oauth.AuthError
	assert.ErrorAs(t, err, &ae)
	assert.Equal(t, "access_denied", ae.Code)
}

func TestAcquireTokenAbortsOnStateMismatch(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", oauth.ErrStateMismatch
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("would-be-pasted\n"))
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	assert.ErrorIs(t, err, oauth.ErrStateMismatch)
}

func TestAcquireTokenAbortsOnContextCancellation(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", context.Canceled
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("would-be-pasted\n"))
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	assert.ErrorIs(t, err, context.Canceled)
}

// --- end-to-end: auth login via OAuth ---

func TestAuthLoginViaOAuthEndToEnd(t *testing.T) {
	isolateConfigHomeForCLI(t)
	forceTTY(t, true)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":981,"email":"acct@example.com"}}}`)
	defer server.Close()

	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "tok-oauth-1", nil
	})

	f := cmdutil.NewFactory("test")
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)

	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader("")) // OAuth path should not consume stdin
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	creds, err := config.LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}
	if assert.Len(t, creds.Contexts, 1) {
		ctx := creds.Contexts[0]
		assert.Equal(t, "production", ctx.Name)
		assert.Equal(t, config.ProductionHost, ctx.Host)
		assert.Equal(t, "tok-oauth-1", ctx.Token, "stored token should come from the OAuth flow")
		assert.Equal(t, "981", ctx.AccountID)
		assert.Equal(t, "alice@example.com", ctx.User)
	}
	assert.Equal(t, "production", creds.ActiveContext)
	assert.Contains(t, stderr.String(), "You're now logged in to DNSimple as alice@example.com")
	assert.Contains(t, stderr.String(), "is now active")
}

func TestAuthLoginViaOAuthOnRollOutErrors(t *testing.T) {
	isolateConfigHomeForCLI(t)
	forceTTY(t, true)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":981,"email":"acct@example.com"}}}`)
	defer server.Close()

	// Rollout window: OAuth is "not provisioned". The command reports the
	// failure and exits instead of falling back to a paste prompt.
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", oauth.ErrNotProvisioned
	})

	f := cmdutil.NewFactory("test")
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)

	cmd.SetIn(strings.NewReader("tok-paste\n")) // must not be consumed
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	err := cmd.RunE(cmd, nil)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "--with-token")

	creds, _ := config.LoadCredentials()
	assert.Empty(t, creds.Contexts, "no context should be stored when browser login fails")
}

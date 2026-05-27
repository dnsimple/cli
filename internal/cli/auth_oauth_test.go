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

func TestAcquireTokenFallsBackToPasteOnErrNotProvisioned(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", oauth.ErrNotProvisioned
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("tok-paste\n"))
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	got, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "tok-paste", got)
	assert.Contains(t, stderr.String(), "not yet available")
}

func TestAcquireTokenPropagatesOAuthErrors(t *testing.T) {
	forceTTY(t, true)
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", fmt.Errorf("oauth: kapow")
	})

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(io.Discard)

	_, err := acquireToken(cmd, &config.Config{}, false)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "kapow")
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
	assert.Contains(t, stderr.String(), "Created context")
}

func TestAuthLoginViaOAuthOnRollOutFallsBackToPaste(t *testing.T) {
	isolateConfigHomeForCLI(t)
	forceTTY(t, true)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":981,"email":"acct@example.com"}}}`)
	defer server.Close()

	// Simulate the rollout window: OAuth is "not provisioned" so the
	// command must fall back to today's paste prompt without erroring out.
	stubLoginViaOAuth(t, func(context.Context, *config.Config, io.Writer) (string, error) {
		return "", oauth.ErrNotProvisioned
	})

	f := cmdutil.NewFactory("test")
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)

	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader("tok-paste\n"))
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	creds, _ := config.LoadCredentials()
	if assert.Len(t, creds.Contexts, 1) {
		assert.Equal(t, "tok-paste", creds.Contexts[0].Token)
	}
	assert.Contains(t, stderr.String(), "not yet available")
}

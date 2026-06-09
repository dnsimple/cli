package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPromptForAccountSelectionReturnsErrorOnEOF(t *testing.T) {
	accountID, err := promptForAccountSelection(strings.NewReader(""), io.Discard, []dnsimple.Account{
		{ID: 101, Email: "one@example.com"},
		{ID: 202, Email: "two@example.com"},
	})

	assert.EqualError(t, err, "no account selected")
	assert.Empty(t, accountID)
}

func TestPromptForAccountSelectionReturnsSelectedAccountID(t *testing.T) {
	accountID, err := promptForAccountSelection(strings.NewReader("2\n"), io.Discard, []dnsimple.Account{
		{ID: 101, Email: "one@example.com"},
		{ID: 202, Email: "two@example.com"},
	})

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "202", accountID)
}

// --- pickAutoContextName ---

func TestPickAutoContextNameUsesBareWhenFree(t *testing.T) {
	creds := &config.Credentials{}
	assert.Equal(t, "production", pickAutoContextName(creds, config.ProductionHost, "981"))
	assert.Equal(t, "sandbox", pickAutoContextName(creds, config.SandboxHost, "24"))
}

func TestPickAutoContextNameFallsThroughToAccountSuffix(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "production", Host: config.ProductionHost, Token: "tok-existing"},
		},
	}
	assert.Equal(t, "production-981", pickAutoContextName(creds, config.ProductionHost, "981"))
}

func TestPickAutoContextNameAppendsNumericSuffix(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "production", Host: config.ProductionHost, Token: "tok-1"},
			{Name: "production-981", Host: config.ProductionHost, Token: "tok-2"},
		},
	}
	assert.Equal(t, "production-981-2", pickAutoContextName(creds, config.ProductionHost, "981"))
}

// --- upsertLoginContext: explicit name ---

func TestUpsertLoginContextExplicitNameFreeCreates(t *testing.T) {
	creds := &config.Credentials{}

	ctx, action, err := upsertLoginContext(creds, config.ProductionHost, "tok-1", "981", "alice@example.com", "personal")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Created", action)
	assert.Equal(t, "personal", ctx.Name)
	assert.Equal(t, "tok-1", ctx.Token)
	assert.Len(t, creds.Contexts, 1)
}

func TestUpsertLoginContextExplicitNameRefreshesOnSameHostAndToken(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "old-account", User: "old@example.com"},
		},
	}

	ctx, action, err := upsertLoginContext(creds, config.ProductionHost, "tok-1", "new-account", "new@example.com", "personal")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Refreshed", action)
	assert.Equal(t, "new-account", ctx.AccountID)
	assert.Equal(t, "new@example.com", ctx.User)
	assert.Len(t, creds.Contexts, 1)
}

func TestUpsertLoginContextExplicitNameRejectsConflictingName(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981"},
		},
	}

	_, _, err := upsertLoginContext(creds, config.ProductionHost, "tok-DIFFERENT", "981", "alice@example.com", "personal")
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), `"personal" already exists`)
	assert.Len(t, creds.Contexts, 1, "no new context should be added on rejection")
}

func TestUpsertLoginContextExplicitNameRejectsTokenStoredElsewhere(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981"},
		},
	}

	_, _, err := upsertLoginContext(creds, config.ProductionHost, "tok-1", "981", "alice@example.com", "alias")
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), `already stored as context "personal"`)
	assert.Len(t, creds.Contexts, 1)
}

// --- upsertLoginContext: auto-derived name ---

func TestUpsertLoginContextAutoDeriveCreatesBareName(t *testing.T) {
	creds := &config.Credentials{}

	ctx, action, err := upsertLoginContext(creds, config.ProductionHost, "tok-1", "981", "alice@example.com", "")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Created", action)
	assert.Equal(t, "production", ctx.Name)
}

func TestUpsertLoginContextAutoDeriveRefreshesOnReLogin(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "production", Host: config.ProductionHost, Token: "tok-1", AccountID: "old", User: "old@example.com"},
		},
	}

	ctx, action, err := upsertLoginContext(creds, config.ProductionHost, "tok-1", "new", "new@example.com", "")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Refreshed", action)
	assert.Equal(t, "production", ctx.Name)
	assert.Equal(t, "new", ctx.AccountID)
	assert.Equal(t, "new@example.com", ctx.User)
	assert.Len(t, creds.Contexts, 1)
}

func TestUpsertLoginContextAutoDeriveAppendsAccountIDOnCollision(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "production", Host: config.ProductionHost, Token: "tok-existing", AccountID: "100"},
		},
	}

	ctx, action, err := upsertLoginContext(creds, config.ProductionHost, "tok-new", "550", "alice@example.com", "")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Created", action)
	assert.Equal(t, "production-550", ctx.Name)
	assert.Len(t, creds.Contexts, 2)
}

// --- auth login command end-to-end ---

func TestAuthLoginCreatesContextAndSetsActive(t *testing.T) {
	isolateConfigHomeForCLI(t)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":981,"email":"acct@example.com"}}}`)
	defer server.Close()

	f := cmdutil.NewFactory("test")
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)

	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader("tok-1\n"))
	cmd.SetErr(&stderr)
	cmd.SetOut(io.Discard)

	if err := cmd.Flags().Set("with-token", "true"); err != nil {
		t.Fatalf("set with-token: %v", err)
	}

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
		assert.Equal(t, "tok-1", ctx.Token)
		assert.Equal(t, "981", ctx.AccountID)
		assert.Equal(t, "alice@example.com", ctx.User)
	}
	assert.Equal(t, "production", creds.ActiveContext)
	assert.Contains(t, stderr.String(), "You're now logged in to DNSimple as alice@example.com")
	assert.Contains(t, stderr.String(), "is now active")
}

func TestAuthLoginWithSandboxFlagCreatesSandboxContext(t *testing.T) {
	isolateConfigHomeForCLI(t)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":24,"email":"acct@example.com"}}}`)
	defer server.Close()

	f := cmdutil.NewFactory("test")
	f.Flags.Sandbox = true
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)

	cmd.SetIn(strings.NewReader("tok-sbx\n"))
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)
	if err := cmd.Flags().Set("with-token", "true"); err != nil {
		t.Fatalf("set with-token: %v", err)
	}

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	creds, _ := config.LoadCredentials()
	if assert.Len(t, creds.Contexts, 1) {
		assert.Equal(t, "sandbox", creds.Contexts[0].Name)
		assert.Equal(t, config.SandboxHost, creds.Contexts[0].Host)
	}
	assert.Equal(t, "sandbox", creds.ActiveContext)
}

func TestAuthLoginWithPersistedSandboxConfigCreatesSandboxContext(t *testing.T) {
	isolateConfigHomeForCLI(t)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":24,"email":"acct@example.com"}}}`)
	defer server.Close()

	f := cmdutil.NewFactory("test")
	f.Config = func() (*config.Config, error) {
		return &config.Config{
			BaseURL: server.URL,
			Sandbox: true,
		}, nil
	}
	cmd := newAuthLoginCmd(f)

	cmd.SetIn(strings.NewReader("tok-sbx\n"))
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)
	if err := cmd.Flags().Set("with-token", "true"); err != nil {
		t.Fatalf("set with-token: %v", err)
	}

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	creds, _ := config.LoadCredentials()
	if assert.Len(t, creds.Contexts, 1) {
		assert.Equal(t, "sandbox", creds.Contexts[0].Name)
		assert.Equal(t, config.SandboxHost, creds.Contexts[0].Host)
	}
	assert.Equal(t, "sandbox", creds.ActiveContext)
}

func TestAuthLoginWithExplicitNameCreatesNamedContext(t *testing.T) {
	isolateConfigHomeForCLI(t)

	server := newWhoamiServer(t, `{"data":{"user":{"id":1,"email":"alice@example.com"},"account":{"id":981,"email":"acct@example.com"}}}`)
	defer server.Close()

	f := cmdutil.NewFactory("test")
	cmd := buildLoginCmdWithBaseURL(t, f, server.URL)
	cmd.SetIn(strings.NewReader("tok-1\n"))
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	if err := cmd.Flags().Set("with-token", "true"); err != nil {
		t.Fatalf("set with-token: %v", err)
	}
	if err := cmd.Flags().Set("name", "personal"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	creds, _ := config.LoadCredentials()
	if assert.Len(t, creds.Contexts, 1) {
		assert.Equal(t, "personal", creds.Contexts[0].Name)
	}
	assert.Equal(t, "personal", creds.ActiveContext)
}

// --- auth logout command end-to-end ---

func TestAuthLogoutRemovesActiveContextWhenNoNameFlag(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1"},
			{Name: "work", Host: config.ProductionHost, Token: "tok-2"},
		},
		ActiveContext: "work",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthLogoutCmd(f)
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	loaded, _ := config.LoadCredentials()
	assert.Len(t, loaded.Contexts, 1)
	assert.Equal(t, "personal", loaded.Contexts[0].Name)
	assert.Equal(t, "personal", loaded.ActiveContext)
}

func TestAuthLogoutRemovesNamedContext(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1"},
			{Name: "work", Host: config.ProductionHost, Token: "tok-2"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthLogoutCmd(f)
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	if err := cmd.Flags().Set("name", "work"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	loaded, _ := config.LoadCredentials()
	assert.Len(t, loaded.Contexts, 1)
	assert.Equal(t, "personal", loaded.Contexts[0].Name)
	assert.Equal(t, "personal", loaded.ActiveContext)
}

func TestAuthLogoutErrorsOnUnknownName(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthLogoutCmd(f)
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	if err := cmd.Flags().Set("name", "missing"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `"missing" not found`)
	}
}

func TestAuthLogoutErrorsWhenNoActiveAndNoName(t *testing.T) {
	isolateConfigHomeForCLI(t)

	f := cmdutil.NewFactory("test")
	cmd := newAuthLogoutCmd(f)
	cmd.SetErr(io.Discard)
	cmd.SetOut(io.Discard)

	err := cmd.RunE(cmd, nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no context to remove")
	}
}

// --- auth list command ---

func TestAuthListPrintsAllContextsAndMarksActive(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981", User: "alice@example.com"},
			{Name: "sandbox", Host: config.SandboxHost, Token: "tok-2", AccountID: "24", User: "bob@example.com"},
		},
		ActiveContext: "sandbox",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthListCmd(f)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	out := stdout.String()
	assert.Contains(t, out, "ENV")
	assert.NotContains(t, out, "ENVIRONMENT")
	assert.Contains(t, out, "personal")
	assert.Contains(t, out, "sandbox")
	assert.Contains(t, out, "production")
	assert.Contains(t, out, "981")
	assert.Contains(t, out, "alice@example.com")
	assert.Contains(t, out, "*", "active context should be marked")
}

func TestAuthListWithNoContextsIsEmpty(t *testing.T) {
	isolateConfigHomeForCLI(t)

	f := cmdutil.NewFactory("test")
	cmd := newAuthListCmd(f)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, stderr.String(), "No contexts")
}

// --- auth switch ---

func TestAuthSwitchByName(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "1111"},
			{Name: "sandbox", Host: config.SandboxHost, Token: "tok-2", AccountID: "2222"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.RunE(cmd, []string{"sandbox"}); !assert.NoError(t, err) {
		return
	}

	loaded, _ := config.LoadCredentials()
	assert.Equal(t, "sandbox", loaded.ActiveContext)
}

func TestAuthSwitchByAccountID(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "1111"},
			{Name: "sandbox", Host: config.SandboxHost, Token: "tok-2", AccountID: "2222"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.RunE(cmd, []string{"2222"}); !assert.NoError(t, err) {
		return
	}

	loaded, _ := config.LoadCredentials()
	assert.Equal(t, "sandbox", loaded.ActiveContext)
}

func TestAuthSwitchByAmbiguousAccountIDErrors(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981"},
			{Name: "work", Host: config.ProductionHost, Token: "tok-2", AccountID: "981"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.RunE(cmd, []string{"981"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "multiple contexts")
		assert.Contains(t, err.Error(), "personal")
		assert.Contains(t, err.Error(), "work")
	}
}

func TestAuthSwitchUnknownErrors(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.RunE(cmd, []string{"missing"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `no context named "missing"`)
	}
}

func TestAuthSwitchToCurrentIsNoOp(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "981"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	var stderr bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&stderr)

	if err := cmd.RunE(cmd, []string{"personal"}); !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, stderr.String(), "Already on context")
}

func TestAuthSwitchEmptyContextsErrors(t *testing.T) {
	isolateConfigHomeForCLI(t)

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.RunE(cmd, []string{"any"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no contexts")
	}
}

func TestAuthSwitchInteractivePicker(t *testing.T) {
	isolateConfigHomeForCLI(t)

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost, Token: "tok-1", AccountID: "1111"},
			{Name: "sandbox", Host: config.SandboxHost, Token: "tok-2", AccountID: "2222"},
		},
		ActiveContext: "personal",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	f := cmdutil.NewFactory("test")
	cmd := newAuthSwitchCmd(f)
	cmd.SetIn(strings.NewReader("2\n"))
	var stderr bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&stderr)

	if err := cmd.RunE(cmd, nil); !assert.NoError(t, err) {
		return
	}

	loaded, _ := config.LoadCredentials()
	assert.Equal(t, "sandbox", loaded.ActiveContext)
	assert.Contains(t, stderr.String(), "env: production, account: 1111")
	assert.Contains(t, stderr.String(), "env: sandbox, account: 2222")
}

// --- promptForContextSelection ---

func TestPromptForContextSelectionInvalidInputErrors(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost},
			{Name: "sandbox", Host: config.SandboxHost},
		},
	}

	_, err := promptForContextSelection(strings.NewReader("99\n"), io.Discard, creds)
	assert.EqualError(t, err, "invalid selection")
}

func TestPromptForContextSelectionEOFErrors(t *testing.T) {
	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "personal", Host: config.ProductionHost},
		},
	}

	_, err := promptForContextSelection(strings.NewReader(""), io.Discard, creds)
	assert.EqualError(t, err, "no context selected")
}

// --- helpers ---

// isolateConfigHomeForCLI redirects HOME to a temp dir so the test reads and
// writes its own credentials.yml without touching the user's config.
func isolateConfigHomeForCLI(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	return home
}

// newWhoamiServer stands up an httptest server that responds to /v2/whoami
// with the given JSON body and to /v2/accounts with an empty list.
func newWhoamiServer(t *testing.T, whoamiJSON string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/whoami":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, whoamiJSON)
		case "/v2/accounts":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[]}`)
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
		}
	}))
}

// buildLoginCmdWithBaseURL constructs the login cobra command and rewires the
// factory's Config so the pre-auth client targets the test server.
func buildLoginCmdWithBaseURL(t *testing.T, f *cmdutil.Factory, baseURL string) *cobra.Command {
	t.Helper()
	cfg := &config.Config{BaseURL: baseURL, Sandbox: f.Flags.Sandbox}
	f.Config = func() (*config.Config, error) { return cfg, nil }
	return newAuthLoginCmd(f)
}

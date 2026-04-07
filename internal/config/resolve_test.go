package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveActiveContext(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100", User: "alice@example.com"},
		},
		ActiveContext: "personal",
	}

	rc, err := Resolve(creds, ResolveOptions{})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "personal", rc.ContextName)
	assert.Equal(t, ProductionHost, rc.Host)
	assert.Equal(t, defaultBaseURL, rc.BaseURL)
	assert.Equal(t, "tok-personal", rc.Token)
	assert.Equal(t, "100", rc.AccountID)
	assert.Equal(t, "alice@example.com", rc.User)
}

func TestResolveContextFlagPicksNamedContext(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100"},
			{Name: "work", Host: ProductionHost, Token: "tok-work", AccountID: "200"},
		},
		ActiveContext: "personal",
	}

	rc, err := Resolve(creds, ResolveOptions{ContextName: "work"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "work", rc.ContextName)
	assert.Equal(t, "tok-work", rc.Token)
	assert.Equal(t, "200", rc.AccountID)
}

func TestResolveContextFlagErrorsOnUnknownName(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok", AccountID: "100"},
		},
		ActiveContext: "personal",
	}

	_, err := Resolve(creds, ResolveOptions{ContextName: "missing"})
	assert.EqualError(t, err, `context "missing" not found`)
}

func TestResolveSandboxFlagPicksLoneSandboxContext(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100"},
			{Name: "sandbox", Host: SandboxHost, Token: "tok-sbx", AccountID: "24"},
		},
		ActiveContext: "personal",
	}

	rc, err := Resolve(creds, ResolveOptions{Sandbox: true})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "sandbox", rc.ContextName)
	assert.Equal(t, SandboxHost, rc.Host)
	assert.Equal(t, sandboxBaseURL, rc.BaseURL)
	assert.Equal(t, "tok-sbx", rc.Token)
	assert.Equal(t, "24", rc.AccountID)
}

func TestResolveSandboxFlagPrefersActiveSandbox(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "sandbox-a", Host: SandboxHost, Token: "tok-a", AccountID: "1"},
			{Name: "sandbox-b", Host: SandboxHost, Token: "tok-b", AccountID: "2"},
		},
		ActiveContext: "sandbox-b",
	}

	rc, err := Resolve(creds, ResolveOptions{Sandbox: true})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "sandbox-b", rc.ContextName)
	assert.Equal(t, "tok-b", rc.Token)
}

func TestResolveSandboxFlagErrorsOnMultipleSandboxContextsWithoutActive(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100"},
			{Name: "sandbox-a", Host: SandboxHost, Token: "tok-a", AccountID: "1"},
			{Name: "sandbox-b", Host: SandboxHost, Token: "tok-b", AccountID: "2"},
		},
		ActiveContext: "personal",
	}

	_, err := Resolve(creds, ResolveOptions{Sandbox: true})
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "multiple sandbox contexts")
	assert.Contains(t, err.Error(), "sandbox-a")
	assert.Contains(t, err.Error(), "sandbox-b")
	assert.Contains(t, err.Error(), "--context")
}

func TestResolveRawFlagsBypassStorage(t *testing.T) {
	creds := &Credentials{} // empty: would otherwise error

	rc, err := Resolve(creds, ResolveOptions{
		Token:   "raw-token",
		Account: "raw-account",
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Empty(t, rc.ContextName)
	assert.Equal(t, "raw-token", rc.Token)
	assert.Equal(t, "raw-account", rc.AccountID)
	assert.Equal(t, ProductionHost, rc.Host)
	assert.Equal(t, defaultBaseURL, rc.BaseURL)
}

func TestResolveRawFlagsWithSandbox(t *testing.T) {
	creds := &Credentials{}

	rc, err := Resolve(creds, ResolveOptions{
		Token:   "raw-token",
		Account: "raw-account",
		Sandbox: true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, SandboxHost, rc.Host)
	assert.Equal(t, sandboxBaseURL, rc.BaseURL)
}

func TestResolveRawFlagsOverrideContext(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "work", Host: ProductionHost, Token: "tok-work", AccountID: "200"},
		},
		ActiveContext: "work",
	}

	rc, err := Resolve(creds, ResolveOptions{
		ContextName: "work",
		Token:       "raw-token",
		Account:     "raw-account",
	})
	if !assert.NoError(t, err) {
		return
	}

	// Context still picked (so ContextName is set), but raw flags win.
	assert.Equal(t, "work", rc.ContextName)
	assert.Equal(t, "raw-token", rc.Token)
	assert.Equal(t, "raw-account", rc.AccountID)
	// Host comes from the context (production), no --sandbox override.
	assert.Equal(t, ProductionHost, rc.Host)
}

func TestResolveContextWithSandboxOverrideHost(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "work", Host: ProductionHost, Token: "tok-work", AccountID: "200"},
		},
		ActiveContext: "work",
	}

	rc, err := Resolve(creds, ResolveOptions{
		ContextName: "work",
		Sandbox:     true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "work", rc.ContextName)
	assert.Equal(t, "tok-work", rc.Token)
	assert.Equal(t, SandboxHost, rc.Host)
	assert.Equal(t, sandboxBaseURL, rc.BaseURL)
}

func TestResolveEnvVarsFillGaps(t *testing.T) {
	t.Setenv("DNSIMPLE_TOKEN", "env-token")
	t.Setenv("DNSIMPLE_ACCOUNT", "env-account")

	creds := &Credentials{}

	rc, err := Resolve(creds, ResolveOptions{})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "env-token", rc.Token)
	assert.Equal(t, "env-account", rc.AccountID)
	assert.Empty(t, rc.ContextName)
}

func TestResolveCfgDefaultAccountUsedWhenNoFlagOrEnv(t *testing.T) {
	t.Setenv("DNSIMPLE_TOKEN", "env-token")
	t.Setenv("DNSIMPLE_ACCOUNT", "")

	creds := &Credentials{}

	rc, err := Resolve(creds, ResolveOptions{DefaultAccount: "cfg-account"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "cfg-account", rc.AccountID)
}

func TestResolveCfgDefaultAccountDoesNotOverrideActiveContext(t *testing.T) {
	t.Setenv("DNSIMPLE_TOKEN", "")
	t.Setenv("DNSIMPLE_ACCOUNT", "")

	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100"},
		},
		ActiveContext: "personal",
	}

	rc, err := Resolve(creds, ResolveOptions{DefaultAccount: "cfg-account"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "personal", rc.ContextName)
	assert.Equal(t, "100", rc.AccountID)
}

func TestResolveCfgDefaultAccountDoesNotOverrideNamedContext(t *testing.T) {
	t.Setenv("DNSIMPLE_TOKEN", "")
	t.Setenv("DNSIMPLE_ACCOUNT", "")

	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-personal", AccountID: "100"},
			{Name: "work", Host: ProductionHost, Token: "tok-work", AccountID: "200"},
		},
		ActiveContext: "personal",
	}

	rc, err := Resolve(creds, ResolveOptions{
		ContextName:    "work",
		DefaultAccount: "cfg-account",
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "work", rc.ContextName)
	assert.Equal(t, "200", rc.AccountID)
}

func TestResolveErrorsWhenNoTokenAvailable(t *testing.T) {
	t.Setenv("DNSIMPLE_TOKEN", "")
	creds := &Credentials{}

	_, err := Resolve(creds, ResolveOptions{Account: "100"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "not authenticated")
	}
}

func TestResolveAllowsMissingAccountForTokenOnlyCommands(t *testing.T) {
	t.Setenv("DNSIMPLE_ACCOUNT", "")
	creds := &Credentials{}

	rc, err := Resolve(creds, ResolveOptions{Token: "tok"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "tok", rc.Token)
	assert.Empty(t, rc.AccountID)
	assert.Equal(t, ProductionHost, rc.Host)
	assert.Equal(t, defaultBaseURL, rc.BaseURL)
}

func TestResolveBaseURLOverrideForTests(t *testing.T) {
	creds := &Credentials{}

	rc, err := Resolve(creds, ResolveOptions{
		Token:           "tok",
		Account:         "100",
		BaseURLOverride: "http://127.0.0.1:9999",
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "http://127.0.0.1:9999", rc.BaseURL)
}

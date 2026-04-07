package cmdutil

import (
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestContextUsesPersistedSandboxSetting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	creds := &config.Credentials{
		Contexts: []*config.Context{
			{Name: "production", Host: config.ProductionHost, Token: "tok-prod", AccountID: "100"},
			{Name: "sandbox", Host: config.SandboxHost, Token: "tok-sbx", AccountID: "24"},
		},
		ActiveContext: "production",
	}
	if err := creds.Save(); err != nil {
		t.Fatalf("save credentials: %v", err)
	}

	f := NewFactory("test")
	f.Config = func() (*config.Config, error) {
		return &config.Config{
			Sandbox: true,
			BaseURL: config.BaseURLForHost(config.SandboxHost),
		}, nil
	}

	rc, err := f.Context()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "sandbox", rc.ContextName)
	assert.Equal(t, config.SandboxHost, rc.Host)
	assert.Equal(t, config.BaseURLForHost(config.SandboxHost), rc.BaseURL)
	assert.Equal(t, "tok-sbx", rc.Token)
	assert.Equal(t, "24", rc.AccountID)
}

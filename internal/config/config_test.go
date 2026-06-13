package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func isolateConfigHome(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
}

func TestLoadDefaults(t *testing.T) {
	isolateConfigHome(t)

	cfg, err := Load()
	if !assert.NoError(t, err) {
		return
	}

	assert.False(t, cfg.Sandbox)
	assert.Equal(t, defaultBaseURL, cfg.BaseURL)
	assert.Empty(t, cfg.DefaultAccount)
	assert.Equal(t, defaultPerPage, cfg.PerPage)
}

func TestLoadFromEnvironment(t *testing.T) {
	isolateConfigHome(t)
	t.Setenv("DNSIMPLE_SANDBOX", "true")
	t.Setenv("DNSIMPLE_DEFAULT_ACCOUNT", "1010")
	t.Setenv("DNSIMPLE_PER_PAGE", "75")

	cfg, err := Load()
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, cfg.Sandbox)
	assert.Equal(t, sandboxBaseURL, cfg.BaseURL)
	assert.Equal(t, "1010", cfg.DefaultAccount)
	assert.Equal(t, 75, cfg.PerPage)
}

func TestSaveAndReload(t *testing.T) {
	isolateConfigHome(t)

	cfg, err := Load()
	if !assert.NoError(t, err) {
		return
	}

	cfg.SetSandbox(true)
	cfg.DefaultAccount = "2020"
	cfg.PerPage = 99

	if !assert.NoError(t, cfg.Save()) {
		return
	}

	reloaded, err := Load()
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, reloaded.Sandbox)
	assert.Equal(t, sandboxBaseURL, reloaded.BaseURL)
	assert.Equal(t, "2020", reloaded.DefaultAccount)
	assert.Equal(t, 99, reloaded.PerPage)
}

func TestSetSandboxUpdatesBaseURL(t *testing.T) {
	cfg := &Config{}

	cfg.SetSandbox(true)
	assert.Equal(t, sandboxBaseURL, cfg.BaseURL)

	cfg.SetSandbox(false)
	assert.Equal(t, defaultBaseURL, cfg.BaseURL)
}

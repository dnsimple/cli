package config

import (
	"testing"
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
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Sandbox {
		t.Fatalf("Sandbox = true, want false")
	}
	if cfg.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, defaultBaseURL)
	}
	if cfg.DefaultAccount != "" {
		t.Fatalf("DefaultAccount = %q, want empty", cfg.DefaultAccount)
	}
	if cfg.PerPage != defaultPerPage {
		t.Fatalf("PerPage = %d, want %d", cfg.PerPage, defaultPerPage)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	isolateConfigHome(t)
	t.Setenv("DNSIMPLE_SANDBOX", "true")
	t.Setenv("DNSIMPLE_DEFAULT_ACCOUNT", "1010")
	t.Setenv("DNSIMPLE_PER_PAGE", "75")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Sandbox {
		t.Fatalf("Sandbox = false, want true")
	}
	if cfg.BaseURL != sandboxBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, sandboxBaseURL)
	}
	if cfg.DefaultAccount != "1010" {
		t.Fatalf("DefaultAccount = %q, want %q", cfg.DefaultAccount, "1010")
	}
	if cfg.PerPage != 75 {
		t.Fatalf("PerPage = %d, want %d", cfg.PerPage, 75)
	}
}

func TestSaveAndReload(t *testing.T) {
	isolateConfigHome(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cfg.SetSandbox(true)
	cfg.DefaultAccount = "2020"
	cfg.PerPage = 99

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save error = %v", err)
	}

	if !reloaded.Sandbox {
		t.Fatalf("Sandbox = false, want true")
	}
	if reloaded.BaseURL != sandboxBaseURL {
		t.Fatalf("BaseURL = %q, want %q", reloaded.BaseURL, sandboxBaseURL)
	}
	if reloaded.DefaultAccount != "2020" {
		t.Fatalf("DefaultAccount = %q, want %q", reloaded.DefaultAccount, "2020")
	}
	if reloaded.PerPage != 99 {
		t.Fatalf("PerPage = %d, want %d", reloaded.PerPage, 99)
	}
}

func TestSetSandboxUpdatesBaseURLAndHostKey(t *testing.T) {
	cfg := &Config{}

	cfg.SetSandbox(true)
	if cfg.BaseURL != sandboxBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, sandboxBaseURL)
	}
	if cfg.HostKey() != "api.sandbox.dnsimple.com" {
		t.Fatalf("HostKey() = %q, want %q", cfg.HostKey(), "api.sandbox.dnsimple.com")
	}

	cfg.SetSandbox(false)
	if cfg.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, defaultBaseURL)
	}
	if cfg.HostKey() != "api.dnsimple.com" {
		t.Fatalf("HostKey() = %q, want %q", cfg.HostKey(), "api.dnsimple.com")
	}
}

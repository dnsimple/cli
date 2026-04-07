package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCredentialsMissingFile(t *testing.T) {
	isolateConfigHome(t)

	creds, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, creds)
	assert.Empty(t, creds.Contexts)
	assert.Empty(t, creds.ActiveContext)
}

func TestCredentialsSaveAndLoad(t *testing.T) {
	isolateConfigHome(t)

	creds := &Credentials{
		Contexts: []*Context{
			{
				Name:      "production",
				Host:      ProductionHost,
				Token:     "token-1",
				AccountID: "1010",
				User:      "user@example.com",
			},
		},
		ActiveContext: "production",
	}

	if !assert.NoError(t, creds.Save()) {
		return
	}

	loaded, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	if assert.Len(t, loaded.Contexts, 1) {
		ctx := loaded.Contexts[0]
		assert.Equal(t, "production", ctx.Name)
		assert.Equal(t, ProductionHost, ctx.Host)
		assert.Equal(t, "token-1", ctx.Token)
		assert.Equal(t, "1010", ctx.AccountID)
		assert.Equal(t, "user@example.com", ctx.User)
	}
	assert.Equal(t, "production", loaded.ActiveContext)
}

func TestCredentialsFindAndActive(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "production", Host: ProductionHost, Token: "tok-prod"},
			{Name: "sandbox", Host: SandboxHost, Token: "tok-sbx"},
		},
		ActiveContext: "sandbox",
	}

	assert.Equal(t, "tok-sbx", creds.Find("sandbox").Token)
	assert.Equal(t, "tok-prod", creds.Find("production").Token)
	assert.Nil(t, creds.Find("missing"))

	active := creds.Active()
	if assert.NotNil(t, active) {
		assert.Equal(t, "sandbox", active.Name)
	}
}

func TestCredentialsActiveReturnsNilWhenUnset(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "production", Host: ProductionHost, Token: "tok-prod"},
		},
	}
	assert.Nil(t, creds.Active())
}

func TestCredentialsAddAppends(t *testing.T) {
	creds := &Credentials{}
	creds.Add(&Context{Name: "personal", Host: ProductionHost, Token: "tok-1"})
	creds.Add(&Context{Name: "work", Host: ProductionHost, Token: "tok-2"})

	assert.Len(t, creds.Contexts, 2)
	assert.Equal(t, "personal", creds.Contexts[0].Name)
	assert.Equal(t, "work", creds.Contexts[1].Name)
}

func TestCredentialsRemove(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-1"},
			{Name: "work", Host: ProductionHost, Token: "tok-2"},
			{Name: "sandbox", Host: SandboxHost, Token: "tok-3"},
		},
		ActiveContext: "work",
	}

	assert.True(t, creds.Remove("work"))
	assert.Len(t, creds.Contexts, 2)
	// Active was removed; should fall back to the first remaining.
	assert.Equal(t, "personal", creds.ActiveContext)

	assert.False(t, creds.Remove("does-not-exist"))
	assert.Len(t, creds.Contexts, 2)
}

func TestCredentialsRemoveLeavesActiveAlone(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-1"},
			{Name: "work", Host: ProductionHost, Token: "tok-2"},
		},
		ActiveContext: "personal",
	}

	creds.Remove("work")
	assert.Equal(t, "personal", creds.ActiveContext)
}

func TestCredentialsRemoveLastClearsActive(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-1"},
		},
		ActiveContext: "personal",
	}

	creds.Remove("personal")
	assert.Empty(t, creds.Contexts)
	assert.Empty(t, creds.ActiveContext)
}

func TestCredentialsSetActive(t *testing.T) {
	creds := &Credentials{
		Contexts: []*Context{
			{Name: "personal", Host: ProductionHost, Token: "tok-1"},
			{Name: "sandbox", Host: SandboxHost, Token: "tok-2"},
		},
	}

	if assert.NoError(t, creds.SetActive("sandbox")) {
		assert.Equal(t, "sandbox", creds.ActiveContext)
	}

	err := creds.SetActive("missing")
	assert.Error(t, err)
}

func TestLoadCredentialsMigratesLegacyHostsFile(t *testing.T) {
	isolateConfigHome(t)

	path := writeLegacyHostsFile(t, `hosts:
    api.dnsimple.com:
        token: prod-token
        account_id: "981"
        user: alice@example.com
    api.sandbox.dnsimple.com:
        token: sbx-token
        account_id: "24"
        user: bob@example.com
`)

	creds, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	if assert.Len(t, creds.Contexts, 2) {
		// Production migrates first because we iterate in a fixed order.
		assert.Equal(t, "production", creds.Contexts[0].Name)
		assert.Equal(t, ProductionHost, creds.Contexts[0].Host)
		assert.Equal(t, "prod-token", creds.Contexts[0].Token)
		assert.Equal(t, "981", creds.Contexts[0].AccountID)
		assert.Equal(t, "alice@example.com", creds.Contexts[0].User)

		assert.Equal(t, "sandbox", creds.Contexts[1].Name)
		assert.Equal(t, SandboxHost, creds.Contexts[1].Host)
		assert.Equal(t, "sbx-token", creds.Contexts[1].Token)
	}
	assert.Equal(t, "production", creds.ActiveContext)

	// File on disk has been rewritten in the contexts shape.
	rewritten, err := os.ReadFile(path)
	if assert.NoError(t, err) {
		assert.Contains(t, string(rewritten), "contexts:")
		assert.NotContains(t, string(rewritten), "hosts:")
	}
}

func TestLoadCredentialsMigratesSandboxOnlyFile(t *testing.T) {
	isolateConfigHome(t)

	writeLegacyHostsFile(t, `hosts:
    api.sandbox.dnsimple.com:
        token: sbx-token
        account_id: "24"
`)

	creds, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	if assert.Len(t, creds.Contexts, 1) {
		assert.Equal(t, "sandbox", creds.Contexts[0].Name)
	}
	assert.Equal(t, "sandbox", creds.ActiveContext)
}

func TestLoadCredentialsMigrationIsIdempotent(t *testing.T) {
	isolateConfigHome(t)

	path := writeLegacyHostsFile(t, `hosts:
    api.dnsimple.com:
        token: prod-token
        account_id: "981"
`)

	// First load triggers migration.
	if _, err := LoadCredentials(); !assert.NoError(t, err) {
		return
	}

	// Capture the on-disk file after the first migration.
	firstAfterMigration, err := os.ReadFile(path)
	if !assert.NoError(t, err) {
		return
	}

	// Second load should be a plain read with no migration side effects.
	creds, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, creds.Contexts, 1)
	assert.Equal(t, "production", creds.ActiveContext)

	// File on disk is unchanged after the second load.
	secondAfterMigration, err := os.ReadFile(path)
	if assert.NoError(t, err) {
		assert.Equal(t, firstAfterMigration, secondAfterMigration)
	}
}

func TestTokenResolutionPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		envToken  string
		flagToken string
		stored    string
		want      string
		wantErr   bool
	}{
		{name: "flag wins over environment and stored", envToken: "env-token", flagToken: "flag-token", stored: "stored-token", want: "flag-token"},
		{name: "environment wins over stored", envToken: "env-token", stored: "stored-token", want: "env-token"},
		{name: "flag wins over stored", flagToken: "flag-token", stored: "stored-token", want: "flag-token"},
		{name: "stored fallback", stored: "stored-token", want: "stored-token"},
		{name: "missing token errors", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateConfigHome(t)
			t.Setenv("DNSIMPLE_TOKEN", tt.envToken)

			cfg := &Config{}
			if tt.stored != "" {
				creds := &Credentials{}
				creds.Set(cfg.HostKey(), &HostCredential{Token: tt.stored})
				if !assert.NoError(t, creds.Save()) {
					return
				}
			}

			got, err := Token(cfg, tt.flagToken)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAccountIDResolutionPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		flagAccount string
		envAccount  string
		cfgAccount  string
		stored      string
		want        string
		wantErr     bool
	}{
		{name: "flag wins over all", flagAccount: "flag-account", envAccount: "env-account", cfgAccount: "cfg-account", stored: "stored-account", want: "flag-account"},
		{name: "environment wins over config and stored", envAccount: "env-account", cfgAccount: "cfg-account", stored: "stored-account", want: "env-account"},
		{name: "config wins over stored", cfgAccount: "cfg-account", stored: "stored-account", want: "cfg-account"},
		{name: "stored fallback", stored: "stored-account", want: "stored-account"},
		{name: "missing account errors", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateConfigHome(t)
			t.Setenv("DNSIMPLE_ACCOUNT", tt.envAccount)

			cfg := &Config{DefaultAccount: tt.cfgAccount}
			if tt.stored != "" {
				creds := &Credentials{}
				creds.Set(cfg.HostKey(), &HostCredential{AccountID: tt.stored})
				if !assert.NoError(t, creds.Save()) {
					return
				}
			}

			got, err := AccountID(cfg, tt.flagAccount)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// writeLegacyHostsFile writes the given YAML content to the credentials path
// inside the isolated config home and returns the path.
func writeLegacyHostsFile(t *testing.T, content string) string {
	t.Helper()

	path, err := credentialsPath()
	if err != nil {
		t.Fatalf("credentialsPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

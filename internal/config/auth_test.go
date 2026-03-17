package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCredentialsMissingFile(t *testing.T) {
	isolateConfigHome(t)

	creds, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, creds.Hosts)
	assert.Len(t, creds.Hosts, 0)
}

func TestCredentialsSaveAndLoad(t *testing.T) {
	isolateConfigHome(t)

	creds := &Credentials{Hosts: map[string]*HostCredential{
		"api.dnsimple.com": {
			Token:     "token-1",
			AccountID: "1010",
			User:      "user@example.com",
		},
	}}

	if !assert.NoError(t, creds.Save()) {
		return
	}

	loaded, err := LoadCredentials()
	if !assert.NoError(t, err) {
		return
	}

	got := loaded.Get("api.dnsimple.com")
	if assert.NotNil(t, got) {
		assert.Equal(t, "token-1", got.Token)
		assert.Equal(t, "1010", got.AccountID)
		assert.Equal(t, "user@example.com", got.User)
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
		{name: "environment wins over flag and stored", envToken: "env-token", flagToken: "flag-token", stored: "stored-token", want: "env-token"},
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
				creds := &Credentials{Hosts: map[string]*HostCredential{
					cfg.HostKey(): {Token: tt.stored},
				}}
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
				creds := &Credentials{Hosts: map[string]*HostCredential{
					cfg.HostKey(): {AccountID: tt.stored},
				}}
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

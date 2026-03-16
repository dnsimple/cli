package config

import "testing"

func TestLoadCredentialsMissingFile(t *testing.T) {
	isolateConfigHome(t)

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if creds.Hosts == nil {
		t.Fatalf("Hosts = nil, want initialized map")
	}
	if len(creds.Hosts) != 0 {
		t.Fatalf("len(Hosts) = %d, want 0", len(creds.Hosts))
	}
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

	if err := creds.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	got := loaded.Get("api.dnsimple.com")
	if got == nil {
		t.Fatalf("Get() = nil, want credential")
	}
	if got.Token != "token-1" || got.AccountID != "1010" || got.User != "user@example.com" {
		t.Fatalf("loaded credential = %#v, want saved values", got)
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
				if err := creds.Save(); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			got, err := Token(cfg, tt.flagToken)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Token() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Token() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Token() = %q, want %q", got, tt.want)
			}
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
				if err := creds.Save(); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			got, err := AccountID(cfg, tt.flagAccount)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("AccountID() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("AccountID() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("AccountID() = %q, want %q", got, tt.want)
			}
		})
	}
}

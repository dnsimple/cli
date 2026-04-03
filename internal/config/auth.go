package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HostCredential stores authentication details for a single API host.
type HostCredential struct {
	Token     string `yaml:"token"`
	AccountID string `yaml:"account_id"`
	User      string `yaml:"user,omitempty"`
}

// Credentials stores authentication details for all known hosts.
type Credentials struct {
	Hosts map[string]*HostCredential `yaml:"hosts"`
}

// credentialsPath returns the path to the credentials file.
func credentialsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFileName+".yml"), nil
}

// LoadCredentials reads stored credentials from disk.
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	creds := &Credentials{Hosts: make(map[string]*HostCredential)}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return creds, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if creds.Hosts == nil {
		creds.Hosts = make(map[string]*HostCredential)
	}

	return creds, nil
}

// Save writes credentials to disk with restricted permissions.
func (c *Credentials) Save() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// Set stores a credential for the given host.
func (c *Credentials) Set(host string, cred *HostCredential) {
	c.Hosts[host] = cred
}

// Get retrieves the credential for the given host.
func (c *Credentials) Get(host string) *HostCredential {
	return c.Hosts[host]
}

// Delete removes the credential for the given host.
func (c *Credentials) Delete(host string) {
	delete(c.Hosts, host)
}

// Token resolves the API token from flag override, environment, or stored credentials.
// The flagToken parameter is the value from --token flag (empty if not set).
func Token(cfg *Config, flagToken string) (string, error) {
	// 1. Flag override
	if flagToken != "" {
		return flagToken, nil
	}

	// 2. Environment variable
	if token := os.Getenv("DNSIMPLE_TOKEN"); token != "" {
		return token, nil
	}

	// 3. Stored credentials
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}

	host := cfg.HostKey()
	cred := creds.Get(host)
	if cred != nil && cred.Token != "" {
		return cred.Token, nil
	}

	return "", fmt.Errorf("not authenticated. Run 'dnsimple auth login' to authenticate")
}

// AccountID resolves the account ID from flag, environment, config, or stored credentials.
func AccountID(cfg *Config, flagAccount string) (string, error) {
	// 1. Flag override
	if flagAccount != "" {
		return flagAccount, nil
	}

	// 2. Environment variable
	if account := os.Getenv("DNSIMPLE_ACCOUNT"); account != "" {
		return account, nil
	}

	// 3. Config default
	if cfg.DefaultAccount != "" {
		return cfg.DefaultAccount, nil
	}

	// 4. Stored credentials
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}

	host := cfg.HostKey()
	cred := creds.Get(host)
	if cred != nil && cred.AccountID != "" {
		return cred.AccountID, nil
	}

	return "", fmt.Errorf("no account specified. Use --account flag, DNSIMPLE_ACCOUNT env var, or run 'dnsimple auth login'")
}

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	productionHost = "api.dnsimple.com"
	sandboxHost    = "api.sandbox.dnsimple.com"
)

// Context is a named authentication context. Each context binds a token to a
// specific (host, account) pair and is identified by a unique name.
type Context struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Token     string `yaml:"token"`
	AccountID string `yaml:"account_id,omitempty"`
	User      string `yaml:"user,omitempty"`
}

// Credentials stores all authentication contexts and the active selection.
type Credentials struct {
	Contexts      []*Context `yaml:"contexts,omitempty"`
	ActiveContext string     `yaml:"active_context,omitempty"`
}

// credentialsFile is the on-disk shape. It supports both the legacy v0
// `hosts:` map and the v1 `contexts:` list so legacy files can be migrated
// in place on first load.
type credentialsFile struct {
	Contexts      []*Context                 `yaml:"contexts,omitempty"`
	ActiveContext string                     `yaml:"active_context,omitempty"`
	Hosts         map[string]*HostCredential `yaml:"hosts,omitempty"`
}

// HostCredential is the legacy per-host credential shape kept for backward
// compatibility with callers that have not yet migrated to the contexts API.
//
// Deprecated: use Context for new code.
type HostCredential struct {
	Token     string `yaml:"token"`
	AccountID string `yaml:"account_id"`
	User      string `yaml:"user,omitempty"`
}

// credentialsPath returns the path to the credentials file.
func credentialsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFileName+".yml"), nil
}

// LoadCredentials reads stored credentials from disk. If the file uses the
// legacy `hosts:` schema, it is migrated in place to the contexts schema.
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, err
	}

	var file credentialsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	creds := &Credentials{
		Contexts:      file.Contexts,
		ActiveContext: file.ActiveContext,
	}

	// Legacy migration: hosts map → contexts list.
	if len(creds.Contexts) == 0 && len(file.Hosts) > 0 {
		creds.migrateFromHosts(file.Hosts)
		if err := creds.save(path); err != nil {
			return nil, err
		}
	}

	return creds, nil
}

// migrateFromHosts converts a legacy `hosts:` map into the contexts list.
// Each host entry becomes a context named after its environment. If both
// production and sandbox exist, production becomes the active context.
func (c *Credentials) migrateFromHosts(hosts map[string]*HostCredential) {
	if h, ok := hosts[productionHost]; ok && h != nil {
		c.Contexts = append(c.Contexts, &Context{
			Name:      "production",
			Host:      productionHost,
			Token:     h.Token,
			AccountID: h.AccountID,
			User:      h.User,
		})
	}
	if h, ok := hosts[sandboxHost]; ok && h != nil {
		c.Contexts = append(c.Contexts, &Context{
			Name:      "sandbox",
			Host:      sandboxHost,
			Token:     h.Token,
			AccountID: h.AccountID,
			User:      h.User,
		})
	}
	if len(c.Contexts) == 0 {
		return
	}
	if _, ok := hosts[productionHost]; ok {
		c.ActiveContext = "production"
	} else {
		c.ActiveContext = c.Contexts[0].Name
	}
}

// Save writes credentials to disk atomically with restricted permissions.
func (c *Credentials) Save() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	return c.save(path)
}

func (c *Credentials) save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Find returns the context with the given name, or nil if not found.
func (c *Credentials) Find(name string) *Context {
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			return ctx
		}
	}
	return nil
}

// Active returns the currently active context, or nil if there is none.
func (c *Credentials) Active() *Context {
	if c.ActiveContext == "" {
		return nil
	}
	return c.Find(c.ActiveContext)
}

// Add appends a context. The caller is responsible for uniqueness — use Find
// first to check whether to mutate an existing entry or add a new one.
func (c *Credentials) Add(ctx *Context) {
	c.Contexts = append(c.Contexts, ctx)
}

// Remove deletes the context with the given name. If it was the active
// context, the active selection is shifted to the first remaining context, or
// cleared if none remain. Returns true if a context was removed.
func (c *Credentials) Remove(name string) bool {
	out := c.Contexts[:0]
	removed := false
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			removed = true
			continue
		}
		out = append(out, ctx)
	}
	c.Contexts = out
	if removed && c.ActiveContext == name {
		c.ActiveContext = ""
		if len(c.Contexts) > 0 {
			c.ActiveContext = c.Contexts[0].Name
		}
	}
	return removed
}

// SetActive marks the named context as active. Returns an error if the name
// does not exist.
func (c *Credentials) SetActive(name string) error {
	if c.Find(name) == nil {
		return fmt.Errorf("context %q not found", name)
	}
	c.ActiveContext = name
	return nil
}

// --- Legacy compatibility shims ---
//
// The Get/Set/Delete methods below operate on the contexts slice using the
// legacy bare-name convention (one context per host, named after the
// environment). They exist so callers that have not yet been rewritten to use
// the contexts API continue to work, and will be removed once those callers
// are migrated.

// Get returns the credentials for the first context whose host matches.
//
// Deprecated: use Find or Active for new code.
func (c *Credentials) Get(host string) *HostCredential {
	for _, ctx := range c.Contexts {
		if ctx.Host == host {
			return &HostCredential{
				Token:     ctx.Token,
				AccountID: ctx.AccountID,
				User:      ctx.User,
			}
		}
	}
	return nil
}

// Set creates or updates a context for the given host using the legacy
// bare-name convention. The resulting context becomes the active one.
//
// Deprecated: use Add for new code.
func (c *Credentials) Set(host string, cred *HostCredential) {
	name := bareNameForHost(host)
	for _, ctx := range c.Contexts {
		if ctx.Host == host {
			ctx.Name = name
			ctx.Token = cred.Token
			ctx.AccountID = cred.AccountID
			ctx.User = cred.User
			c.ActiveContext = name
			return
		}
	}
	c.Contexts = append(c.Contexts, &Context{
		Name:      name,
		Host:      host,
		Token:     cred.Token,
		AccountID: cred.AccountID,
		User:      cred.User,
	})
	c.ActiveContext = name
}

// Delete removes all contexts whose host matches.
//
// Deprecated: use Remove for new code.
func (c *Credentials) Delete(host string) {
	out := c.Contexts[:0]
	for _, ctx := range c.Contexts {
		if ctx.Host != host {
			out = append(out, ctx)
		}
	}
	c.Contexts = out
	if c.ActiveContext != "" && c.Find(c.ActiveContext) == nil {
		c.ActiveContext = ""
		if len(c.Contexts) > 0 {
			c.ActiveContext = c.Contexts[0].Name
		}
	}
}

func bareNameForHost(host string) string {
	if host == sandboxHost {
		return "sandbox"
	}
	return "production"
}

// Token resolves the API token from flag override, environment, or stored
// credentials. The flagToken parameter is the value from the --token flag
// (empty if not set).
func Token(cfg *Config, flagToken string) (string, error) {
	if flagToken != "" {
		return flagToken, nil
	}
	if token := os.Getenv("DNSIMPLE_TOKEN"); token != "" {
		return token, nil
	}

	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}

	cred := creds.Get(cfg.HostKey())
	if cred != nil && cred.Token != "" {
		return cred.Token, nil
	}

	return "", fmt.Errorf("not authenticated. Run 'dnsimple auth login' to authenticate")
}

// AccountID resolves the account ID from flag, environment, config, or
// stored credentials.
func AccountID(cfg *Config, flagAccount string) (string, error) {
	if flagAccount != "" {
		return flagAccount, nil
	}
	if account := os.Getenv("DNSIMPLE_ACCOUNT"); account != "" {
		return account, nil
	}
	if cfg.DefaultAccount != "" {
		return cfg.DefaultAccount, nil
	}

	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}

	cred := creds.Get(cfg.HostKey())
	if cred != nil && cred.AccountID != "" {
		return cred.AccountID, nil
	}

	return "", fmt.Errorf("no account specified. Use --account flag, DNSIMPLE_ACCOUNT env var, or run 'dnsimple auth login'")
}

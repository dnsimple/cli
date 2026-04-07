package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// ProductionHost is the canonical hostname of the production DNSimple API.
	ProductionHost = "api.dnsimple.com"
	// SandboxHost is the canonical hostname of the sandbox DNSimple API.
	SandboxHost = "api.sandbox.dnsimple.com"
)

// EnvironmentName returns a human-readable environment label for a host.
// Recognised hosts return "production" or "sandbox"; unknown hosts return
// the host string verbatim.
func EnvironmentName(host string) string {
	switch host {
	case ProductionHost:
		return "production"
	case SandboxHost:
		return "sandbox"
	default:
		return host
	}
}

// HostForSandbox returns the canonical host for the given environment toggle.
func HostForSandbox(sandbox bool) string {
	if sandbox {
		return SandboxHost
	}
	return ProductionHost
}

// BaseURLForHost returns the canonical API URL for a known host. Unknown
// hosts fall back to the production URL.
func BaseURLForHost(host string) string {
	if host == SandboxHost {
		return sandboxBaseURL
	}
	return defaultBaseURL
}

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
	if h, ok := hosts[ProductionHost]; ok && h != nil {
		c.Contexts = append(c.Contexts, &Context{
			Name:      "production",
			Host:      ProductionHost,
			Token:     h.Token,
			AccountID: h.AccountID,
			User:      h.User,
		})
	}
	if h, ok := hosts[SandboxHost]; ok && h != nil {
		c.Contexts = append(c.Contexts, &Context{
			Name:      "sandbox",
			Host:      SandboxHost,
			Token:     h.Token,
			AccountID: h.AccountID,
			User:      h.User,
		})
	}
	if len(c.Contexts) == 0 {
		return
	}
	if _, ok := hosts[ProductionHost]; ok {
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


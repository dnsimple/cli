package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultBaseURL        = "https://api.dnsimple.com"
	sandboxBaseURL        = "https://api.sandbox.dnsimple.com"
	defaultPerPage        = 30
	configDirName         = "dnsimple"
	configFileName        = "config"
	credentialsFileName   = "credentials"

	// baseURLEnvVar overrides the API base URL for local development against
	// a local API server. Consulted by both Load/SetSandbox (the
	// pre-auth `auth login` path) and Resolve (authenticated commands) so the
	// override moves every API call, including the OAuth token endpoint.
	baseURLEnvVar = "DNSIMPLE_BASE_URL"
)

// Config holds the CLI configuration.
type Config struct {
	v *viper.Viper

	// BaseURL is the API base URL (production or sandbox).
	BaseURL string

	// Sandbox indicates whether to use the sandbox environment.
	Sandbox bool

	// DefaultAccount is the default account ID.
	DefaultAccount string

	// PerPage is the default number of items per page for list commands.
	PerPage int

	// OAuthLogin opts `auth login` into the interactive browser flow; off by
	// default during the dark-launch rollout (see --web / DNSIMPLE_OAUTH_LOGIN).
	OAuthLogin bool
}

// Dir returns the configuration directory path.
func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configDirName), nil
}

// Load reads the configuration from the config file and environment.
func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigName(configFileName)
	v.SetConfigType("yml")
	v.AddConfigPath(dir)

	v.SetEnvPrefix("DNSIMPLE")
	v.AutomaticEnv()

	v.SetDefault("sandbox", false)
	v.SetDefault("per_page", defaultPerPage)
	v.SetDefault("default_account", "")
	v.SetDefault("oauth_login", false)

	// Config file is optional
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	cfg := &Config{
		v:              v,
		Sandbox:        v.GetBool("sandbox"),
		DefaultAccount: v.GetString("default_account"),
		PerPage:        v.GetInt("per_page"),
		OAuthLogin:     v.GetBool("oauth_login"),
	}

	cfg.BaseURL = baseURLForSandbox(cfg.Sandbox)

	return cfg, nil
}

// SetSandbox overrides the sandbox setting (from --sandbox flag).
func (c *Config) SetSandbox(sandbox bool) {
	c.Sandbox = sandbox
	c.BaseURL = baseURLForSandbox(sandbox)
}

// baseURLForSandbox returns the API base URL for the given environment,
// honoring DNSIMPLE_BASE_URL when set. The env override wins over the
// production/sandbox default, mirroring the precedence in Resolve so the
// `auth login` flow (which builds a Config, not a ResolvedContext, because no
// token exists yet) targets the same local host as authenticated commands.
func baseURLForSandbox(sandbox bool) string {
	if v := strings.TrimSpace(os.Getenv(baseURLEnvVar)); v != "" {
		return v
	}
	if sandbox {
		return sandboxBaseURL
	}
	return defaultBaseURL
}

// Save writes the current configuration to disk.
func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	c.v.Set("sandbox", c.Sandbox)
	c.v.Set("default_account", c.DefaultAccount)
	c.v.Set("per_page", c.PerPage)
	c.v.Set("oauth_login", c.OAuthLogin)

	return c.v.WriteConfigAs(filepath.Join(dir, configFileName+".yml"))
}


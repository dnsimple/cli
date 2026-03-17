package cmdutil

import (
	"github.com/dnsimple/dnsimple-cli/internal/client"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-cli/internal/output"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// Factory provides lazy access to shared dependencies for commands.
type Factory struct {
	// Version is the CLI version used for downstream integrations such as the API user agent.
	Version string

	// Config returns the loaded configuration.
	Config func() (*config.Config, error)

	// Client returns a configured DNSimple API client.
	// It resolves the token and builds the client lazily.
	Client func() (*dnsimple.Client, error)

	// Printer returns the output printer configured with the current format settings.
	Printer func(cmd *cobra.Command) *output.Printer

	// AccountID resolves the account ID from flags, env, config, or credentials.
	AccountID func() (string, error)

	// Flags holds the raw global flag values for resolution.
	Flags *GlobalFlags
}

// GlobalFlags holds the values of global flags set on the root command.
type GlobalFlags struct {
	Account string
	Token   string
	Sandbox bool
	JSON    bool
	Format  string
	NoColor bool
	Debug   bool
	Quiet   bool
}

// NewFactory creates a new Factory with lazy initialization.
func NewFactory(version string) *Factory {
	flags := &GlobalFlags{}
	var cachedConfig *config.Config
	var cachedClient *dnsimple.Client

	f := &Factory{
		Version: version,
		Flags:   flags,
	}

	f.Config = func() (*config.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		// Apply flag overrides
		if flags.Sandbox {
			cfg.SetSandbox(true)
		}
		cachedConfig = cfg
		return cfg, nil
	}

	f.Client = func() (*dnsimple.Client, error) {
		if cachedClient != nil {
			return cachedClient, nil
		}
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		token, err := config.Token(cfg, flags.Token)
		if err != nil {
			return nil, err
		}
		c := client.NewClient(cfg, token, f.Version)
		if flags.Debug {
			c.Debug = true
		}
		cachedClient = c
		return c, nil
	}

	f.Printer = func(cmd *cobra.Command) *output.Printer {
		format := output.FormatTable
		if flags.JSON {
			format = output.FormatJSON
		} else if flags.Format != "" {
			format = output.FormatTemplate
		}
		return output.NewPrinter(cmd.OutOrStdout(), cmd.ErrOrStderr(), format, flags.Format, flags.NoColor)
	}

	f.AccountID = func() (string, error) {
		cfg, err := f.Config()
		if err != nil {
			return "", err
		}
		return config.AccountID(cfg, flags.Account)
	}

	return f
}

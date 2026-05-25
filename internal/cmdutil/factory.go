package cmdutil

import (
	"errors"

	"github.com/dnsimple/cli/internal/client"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/cli/internal/output"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

// Factory provides lazy access to shared dependencies for commands.
type Factory struct {
	// Version is the CLI version used for downstream integrations such as the API user agent.
	Version string

	// Config returns the loaded configuration.
	Config func() (*config.Config, error)

	// Context returns the resolved authentication context for this invocation.
	// Lazy and cached: the first call performs resolution; subsequent calls
	// return the same value.
	Context func() (*config.ResolvedContext, error)

	// Client returns a configured DNSimple API client.
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
	Context string
	Token   string
	Sandbox bool
	JSON    bool
	Format  string
	NoColor bool
	Debug   bool
}

// NewFactory creates a new Factory with lazy initialization.
func NewFactory(version string) *Factory {
	flags := &GlobalFlags{}
	var cachedConfig *config.Config
	var cachedContext *config.ResolvedContext
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

	f.Context = func() (*config.ResolvedContext, error) {
		if cachedContext != nil {
			return cachedContext, nil
		}
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		creds, err := config.LoadCredentials()
		if err != nil {
			return nil, err
		}
		rc, err := config.Resolve(creds, config.ResolveOptions{
			Token:          flags.Token,
			Account:        flags.Account,
			ContextName:    flags.Context,
			Sandbox:        cfg.Sandbox,
			DefaultAccount: cfg.DefaultAccount,
		})
		if err != nil {
			return nil, err
		}
		cachedContext = rc
		return rc, nil
	}

	f.Client = func() (*dnsimple.Client, error) {
		if cachedClient != nil {
			return cachedClient, nil
		}
		rc, err := f.Context()
		if err != nil {
			return nil, err
		}
		cachedClient = client.New(client.Options{
			BaseURL: rc.BaseURL,
			Token:   rc.Token,
			Version: f.Version,
			Debug:   flags.Debug,
		})
		return cachedClient, nil
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
		rc, err := f.Context()
		if err != nil {
			return "", err
		}
		if rc.AccountID == "" {
			return "", errors.New("no account specified. Use --account flag, DNSIMPLE_ACCOUNT env var, or run 'dnsimple auth login'")
		}
		return rc.AccountID, nil
	}

	return f
}

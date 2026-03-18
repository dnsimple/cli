package cli

import (
	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// buildRootCmd creates the root command with all subcommands and global flags.
func buildRootCmd(f *cmdutil.Factory) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "dnsimple",
		Short:         "DNSimple CLI",
		Long:          "Work with DNSimple from the command line.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&f.Flags.Account, "account", "a", "", "Account ID to operate on")
	rootCmd.PersistentFlags().StringVar(&f.Flags.Token, "token", "", "API token (overrides DNSIMPLE_TOKEN env var)")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.Sandbox, "sandbox", false, "Use sandbox environment")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.JSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&f.Flags.Format, "format", "", "Custom output format (Go template)")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.Debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&f.Flags.Quiet, "quiet", "q", false, "Suppress non-essential output")

	// Register commands
	rootCmd.AddCommand(newAuthCmd(f))
	rootCmd.AddCommand(newWhoamiCmd(f))
	rootCmd.AddCommand(newAccountsCmd(f))
	rootCmd.AddCommand(newDomainsCmd(f))
	rootCmd.AddCommand(newZonesCmd(f))
	rootCmd.AddCommand(newRecordsCmd(f))
	rootCmd.AddCommand(newRegistrarCmd(f))
	rootCmd.AddCommand(newContactsCmd(f))
	rootCmd.AddCommand(newCertificatesCmd(f))
	rootCmd.AddCommand(newServicesCmd(f))
	rootCmd.AddCommand(newTemplatesCmd(f))
	rootCmd.AddCommand(newTldsCmd(f))
	rootCmd.AddCommand(newWebhooksCmd(f))
	rootCmd.AddCommand(newVanityNameServersCmd(f))
	rootCmd.AddCommand(newBillingCmd(f))
	rootCmd.AddCommand(newAnalyticsCmd(f))
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newAICmd())

	return rootCmd
}

// Execute runs the CLI with the given version and arguments.
func Execute(version string, args []string) int {
	f := cmdutil.NewFactory(version)
	rootCmd := buildRootCmd(f)
	rootCmd.Version = version
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		cmdutil.FormatAPIError(rootCmd.ErrOrStderr(), err)
		return cmdutil.ExitError
	}

	return cmdutil.ExitOK
}

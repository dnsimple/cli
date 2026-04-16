package cli

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/update"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	rootCmd.PersistentFlags().StringVar(&f.Flags.Context, "context", "", "Authentication context to use for this invocation (overrides the active context)")
	rootCmd.PersistentFlags().StringVar(&f.Flags.Token, "token", "", "API token (overrides DNSIMPLE_TOKEN env var)")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.Sandbox, "sandbox", false, "Use sandbox environment")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.JSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&f.Flags.Format, "format", "", "Custom output format (Go template over the resource; use {{.Field}} for single items or {{range .}}{{.Field}}{{end}} for lists)")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&f.Flags.Debug, "debug", false, "Enable debug logging")

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
	rootCmd.AddCommand(newResearchCmd(f))
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

	// Start async update check.
	debug := containsFlag(args, "--debug")
	var updateCh <-chan *update.CheckResult
	if update.ShouldCheck(update.Opts{
		CurrentVersion: version,
		IsTerminal:     term.IsTerminal(int(os.Stderr.Fd())),
		Debug:          debug,
		Args:           args,
	}) {
		updateCh = update.CheckAsync(context.Background(), version, debug)
	}

	exitCode := cmdutil.ExitOK
	if err := rootCmd.Execute(); err != nil {
		cmdutil.FormatAPIError(rootCmd.ErrOrStderr(), err)
		exitCode = cmdutil.ExitError
	}

	// Print update notice if the check completed.
	if updateCh != nil {
		select {
		case result := <-updateCh:
			update.PrintNotice(os.Stderr, result)
		case <-time.After(2 * time.Second):
		}
	}

	return exitCode
}

// containsFlag returns true if any of the given flags appear in args.
func containsFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			if arg == flag || strings.HasPrefix(arg, flag+"=") {
				return true
			}
		}
	}
	return false
}

package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for dnsimple CLI.

To load completions:

  bash:
    source <(dnsimple completion bash)

  zsh:
    # If shell completion is not already enabled in your environment, enable it:
    echo "autoload -U compinit; compinit" >> ~/.zshrc

    # Generate and load completions:
    dnsimple completion zsh > "${fpath[1]}/_dnsimple"

  fish:
    dnsimple completion fish | source

  powershell:
    dnsimple completion powershell | Out-String | Invoke-Expression`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate the autocompletion script for bash",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate the autocompletion script for zsh",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "fish",
		Short: "Generate the autocompletion script for fish",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "powershell",
		Short: "Generate the autocompletion script for powershell",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		},
	})

	return cmd
}

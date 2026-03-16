package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
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
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return cmd.Help()
			}
		},
	}

	return cmd
}

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const destructiveYesFlagHelp = "Confirm the destructive operation without prompting"

func addYesFlag(cmd *cobra.Command, yes *bool) {
	cmd.Flags().BoolVarP(yes, "yes", "y", false, destructiveYesFlagHelp)
}

func confirmDestructiveAction(cmd *cobra.Command, yes bool, prompt string) error {
	return confirmDestructiveActionInput(cmd.InOrStdin(), cmd.ErrOrStderr(), isInteractiveInput(cmd.InOrStdin()), yes, prompt)
}

func confirmDestructiveActionInput(in io.Reader, errOut io.Writer, interactive bool, yes bool, prompt string) error {
	if yes {
		return nil
	}
	if !interactive {
		return fmt.Errorf("destructive operation requires confirmation; rerun with --yes")
	}

	fmt.Fprintf(errOut, "%s [y/N]: ", prompt)

	answer, err := scanLine(in)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("confirmation declined")
	}
}

func isInteractiveInput(in io.Reader) bool {
	f, ok := in.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

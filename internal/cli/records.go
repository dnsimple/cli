package cli

import (
	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newRecordsCmd creates a top-level "records" alias that delegates to "zones records".
func newRecordsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "records",
		Short:       "Manage zone records (alias for 'zones records')",
		Annotations: map[string]string{aiAliasTargetAnnotation: "dnsimple zones records"},
	}

	cmd.AddCommand(newRecordsListCmd(f))
	cmd.AddCommand(newRecordsGetCmd(f))
	cmd.AddCommand(newRecordsCreateCmd(f))
	cmd.AddCommand(newRecordsUpdateCmd(f))
	cmd.AddCommand(newRecordsDeleteCmd(f))
	cmd.AddCommand(newRecordsCheckDistributionCmd(f))

	return cmd
}

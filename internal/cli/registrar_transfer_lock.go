package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// transferLockOutput adapts DomainTransferLock for output.
type transferLockOutput struct {
	Data *dnsimple.DomainTransferLock `json:"data"`
}

func (t *transferLockOutput) TableHeaders() []string {
	return []string{"ENABLED"}
}

func (t *transferLockOutput) TableRows() [][]string {
	return [][]string{{strconv.FormatBool(t.Data.Enabled)}}
}

func (t *transferLockOutput) JSONData() any { return t }

func newRegistrarTransferLockCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-lock",
		Short: "Manage domain transfer lock",
	}

	cmd.AddCommand(newTransferLockStatusCmd(f))
	cmd.AddCommand(newTransferLockEnableCmd(f))
	cmd.AddCommand(newTransferLockDisableCmd(f))

	return cmd
}

func newTransferLockStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status <domain>",
		Short: "Get transfer lock status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Registrar.GetDomainTransferLock(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer().Print(&transferLockOutput{Data: resp.Data})
		},
	}
}

func newTransferLockEnableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable transfer lock",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			_, err = c.Registrar.EnableDomainTransferLock(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Transfer lock enabled for %s\n", args[0])
			}
			return nil
		},
	}
}

func newTransferLockDisableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable transfer lock",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			_, err = c.Registrar.DisableDomainTransferLock(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Transfer lock disabled for %s\n", args[0])
			}
			return nil
		},
	}
}

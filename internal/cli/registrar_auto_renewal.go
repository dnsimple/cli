package cli

import (
	"context"
	"fmt"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newRegistrarAutoRenewalCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-renewal",
		Short: "Manage domain auto-renewal",
	}

	cmd.AddCommand(newAutoRenewalEnableCmd(f))
	cmd.AddCommand(newAutoRenewalDisableCmd(f))

	return cmd
}

func newAutoRenewalEnableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable auto-renewal",
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

			_, err = c.Registrar.EnableDomainAutoRenewal(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Auto-renewal enabled for %s\n", args[0])
			}
			return nil
		},
	}
}

func newAutoRenewalDisableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable auto-renewal",
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

			_, err = c.Registrar.DisableDomainAutoRenewal(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Auto-renewal disabled for %s\n", args[0])
			}
			return nil
		},
	}
}

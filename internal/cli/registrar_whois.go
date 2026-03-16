package cli

import (
	"context"
	"fmt"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newRegistrarWhoisPrivacyCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whois-privacy",
		Short: "Manage WHOIS privacy",
	}

	cmd.AddCommand(newWhoisPrivacyEnableCmd(f))
	cmd.AddCommand(newWhoisPrivacyDisableCmd(f))

	return cmd
}

func newWhoisPrivacyEnableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable WHOIS privacy",
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

			_, err = c.Registrar.EnableWhoisPrivacy(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "WHOIS privacy enabled for %s\n", args[0])
			}
			return nil
		},
	}
}

func newWhoisPrivacyDisableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable WHOIS privacy",
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

			_, err = c.Registrar.DisableWhoisPrivacy(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "WHOIS privacy disabled for %s\n", args[0])
			}
			return nil
		},
	}
}

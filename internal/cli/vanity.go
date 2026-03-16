package cli

import (
	"context"
	"fmt"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newVanityNameServersCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vanity-name-servers",
		Short: "Manage vanity name servers",
	}

	cmd.AddCommand(newVanityEnableCmd(f))
	cmd.AddCommand(newVanityDisableCmd(f))

	return cmd
}

func newVanityEnableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable vanity name servers",
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

			_, err = c.VanityNameServers.EnableVanityNameServers(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Vanity name servers enabled for %s\n", args[0])
			}
			return nil
		},
	}
}

func newVanityDisableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable vanity name servers",
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

			_, err = c.VanityNameServers.DisableVanityNameServers(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Vanity name servers disabled for %s\n", args[0])
			}
			return nil
		},
	}
}

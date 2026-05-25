package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

// dnssecOutput adapts Dnssec for output.
type dnssecOutput struct {
	Data *dnsimple.Dnssec `json:"data"`
}

func (d *dnssecOutput) TableHeaders() []string {
	return []string{"ENABLED"}
}

func (d *dnssecOutput) TableRows() [][]string {
	return [][]string{{strconv.FormatBool(d.Data.Enabled)}}
}

func (d *dnssecOutput) JSONData() any { return d }

func (d *dnssecOutput) TemplateData() any { return d.Data }

func newDomainsDnssecCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dnssec",
		Short: "Manage DNSSEC for a domain",
	}

	cmd.AddCommand(newDnssecStatusCmd(f))
	cmd.AddCommand(newDnssecEnableCmd(f))
	cmd.AddCommand(newDnssecDisableCmd(f))

	return cmd
}

func newDnssecStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status <domain>",
		Short: "Get DNSSEC status",
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

			resp, err := c.Domains.GetDnssec(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&dnssecOutput{Data: resp.Data})
		},
	}
}

func newDnssecEnableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable DNSSEC",
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

			_, err = c.Domains.EnableDnssec(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "DNSSEC enabled for %s\n", args[0])
			return nil
		},
	}
}

func newDnssecDisableCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable DNSSEC",
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

			_, err = c.Domains.DisableDnssec(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "DNSSEC disabled for %s\n", args[0])
			return nil
		},
	}
}

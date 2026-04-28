package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// delegationOutput adapts Delegation for output.
type delegationOutput struct {
	Data *dnsimple.Delegation `json:"data"`
}

func (d *delegationOutput) TableHeaders() []string {
	return []string{"NAME SERVER"}
}

func (d *delegationOutput) TableRows() [][]string {
	if d.Data == nil {
		return nil
	}
	rows := make([][]string, len(*d.Data))
	for i, ns := range *d.Data {
		rows[i] = []string{ns}
	}
	return rows
}

func (d *delegationOutput) JSONData() any { return d }

func (d *delegationOutput) TemplateData() any {
	if d.Data == nil {
		return nil
	}
	return *d.Data
}

func newRegistrarDelegationCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation",
		Short: "Manage domain name server delegation",
	}

	cmd.AddCommand(newDelegationGetCmd(f))
	cmd.AddCommand(newDelegationChangeCmd(f))

	return cmd
}

func newDelegationGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <domain>",
		Short: "Get current name servers",
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

			resp, err := c.Registrar.GetDomainDelegation(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&delegationOutput{Data: resp.Data})
		},
	}
}

func newDelegationChangeCmd(f *cmdutil.Factory) *cobra.Command {
	var nameServers []string

	cmd := &cobra.Command{
		Use:   "change <domain>",
		Short: "Change name server delegation",
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

			delegation := dnsimple.Delegation(nameServers)
			resp, err := c.Registrar.ChangeDomainDelegation(context.Background(), accountID, args[0], &delegation)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Delegation updated for %s: %s\n", args[0], strings.Join(nameServers, ", "))
			_ = resp
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&nameServers, "ns", nil, "Name servers (comma-separated)")
	_ = cmd.MarkFlagRequired("ns")

	return cmd
}

package cli

import (
	"context"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type accountList struct {
	Data []dnsimple.Account `json:"data"`
}

func (a *accountList) TableHeaders() []string {
	return []string{"ID", "EMAIL", "PLAN"}
}

func (a *accountList) TableRows() [][]string {
	rows := make([][]string, len(a.Data))
	for i, acct := range a.Data {
		rows[i] = []string{
			strconv.FormatInt(acct.ID, 10),
			acct.Email,
			acct.PlanIdentifier,
		}
	}
	return rows
}

func (a *accountList) JSONData() any {
	return a
}

func newAccountsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts",
	}

	cmd.AddCommand(newAccountsListCmd(f))
	return cmd
}

func newAccountsListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Accounts.ListAccounts(context.Background(), nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&accountList{Data: resp.Data})
		},
	}
}

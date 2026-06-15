package cli

import (
	"context"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

type chargeList struct {
	Data       []dnsimple.Charge    `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (c *chargeList) TableHeaders() []string {
	return []string{"INVOICED AT", "TOTAL", "BALANCE", "REFERENCE", "STATE"}
}

func (c *chargeList) TableRows() [][]string {
	rows := make([][]string, len(c.Data))
	for i, ch := range c.Data {
		rows[i] = []string{
			ch.InvoicedAt,
			ch.TotalAmount.String(),
			ch.BalanceAmount.String(),
			ch.Reference,
			ch.State,
		}
	}
	return rows
}

func (c *chargeList) JSONData() any { return c }

func (c *chargeList) TemplateData() any { return c.Data }

func newBillingCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing and charges",
	}

	cmd.AddCommand(newBillingChargesCmd(f))

	return cmd
}

func newBillingChargesCmd(f *cmdutil.Factory) *cobra.Command {
	var startDate, endDate, sort string

	cmd := &cobra.Command{
		Use:   "charges",
		Short: "List billing charges",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			opts := dnsimple.ListChargesOptions{
				StartDate: startDate,
				EndDate:   endDate,
				Sort:      sort,
			}

			resp, err := c.Billing.ListCharges(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).PrintList(&chargeList{Data: resp.Data, Pagination: resp.Pagination}, pageHint(cmd, resp.Pagination, len(resp.Data), "charges"))
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")

	return cmd
}

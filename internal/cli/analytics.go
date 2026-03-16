package cli

import (
	"context"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type analyticsOutput struct {
	Data       []dnsimple.DnsAnalytics `json:"data"`
	Pagination *dnsimple.Pagination    `json:"pagination,omitempty"`
}

func (a *analyticsOutput) TableHeaders() []string {
	return []string{"ZONE", "DATE", "VOLUME"}
}

func (a *analyticsOutput) TableRows() [][]string {
	rows := make([][]string, len(a.Data))
	for i, d := range a.Data {
		rows[i] = []string{
			d.ZoneName,
			d.Date,
			strconv.FormatInt(d.Volume, 10),
		}
	}
	return rows
}

func (a *analyticsOutput) JSONData() any { return a }

func newAnalyticsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "DNS analytics",
	}

	cmd.AddCommand(newAnalyticsQueryCmd(f))

	return cmd
}

func newAnalyticsQueryCmd(f *cmdutil.Factory) *cobra.Command {
	var startDate, endDate, groupings, sort string
	var page, perPage int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query DNS analytics",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountIDStr, err := f.AccountID()
			if err != nil {
				return err
			}

			accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
			if err != nil {
				return err
			}

			opts := &dnsimple.DnsAnalyticsOptions{}
			if startDate != "" {
				opts.StartDate = &startDate
			}
			if endDate != "" {
				opts.EndDate = &endDate
			}
			if groupings != "" {
				opts.Groupings = &groupings
			}
			if sort != "" {
				opts.Sort = &sort
			}
			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}

			resp, err := c.DnsAnalytics.Query(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer().Print(&analyticsOutput{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO8601)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO8601)")
	cmd.Flags().StringVar(&groupings, "groupings", "", "Group by (comma-separated)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")

	return cmd
}

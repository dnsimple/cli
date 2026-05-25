package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// supportedGroupings lists the grouping keys the CLI accepts for the
// analytics query command. The values match the column headers the
// DNSimple API returns and that the dnsimple-go client knows how to
// unmarshal.
var supportedGroupings = []string{"zone_name", "date"}

// normalizeGroupings parses the comma-separated --groupings flag value,
// validates each key against supportedGroupings and removes duplicates
// while preserving the order in which the user listed them. It returns
// the ordered list of grouping keys along with the canonical
// comma-separated string suitable for the API.
func normalizeGroupings(raw string) ([]string, string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, "", nil
	}

	supported := make(map[string]bool, len(supportedGroupings))
	for _, key := range supportedGroupings {
		supported[key] = true
	}

	parts := strings.Split(raw, ",")
	seen := make(map[string]bool, len(parts))
	ordered := make([]string, 0, len(parts))

	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		if !supported[key] {
			return nil, "", fmt.Errorf("unsupported grouping %q (supported: %s)", key, strings.Join(supportedGroupings, ", "))
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		ordered = append(ordered, key)
	}

	return ordered, strings.Join(ordered, ","), nil
}

type analyticsOutput struct {
	Data       []dnsimple.DnsAnalytics `json:"data"`
	Pagination *dnsimple.Pagination    `json:"pagination,omitempty"`
	Groupings  []string                `json:"-"`
}

func (a *analyticsOutput) TableHeaders() []string {
	headers := make([]string, 0, len(a.Groupings)+1)
	for _, g := range a.Groupings {
		switch g {
		case "zone_name":
			headers = append(headers, "ZONE")
		case "date":
			headers = append(headers, "DATE")
		default:
			headers = append(headers, strings.ToUpper(g))
		}
	}
	headers = append(headers, "VOLUME")
	return headers
}

func (a *analyticsOutput) TableRows() [][]string {
	rows := make([][]string, len(a.Data))
	for i, d := range a.Data {
		row := make([]string, 0, len(a.Groupings)+1)
		for _, g := range a.Groupings {
			switch g {
			case "zone_name":
				row = append(row, d.ZoneName)
			case "date":
				row = append(row, d.Date)
			default:
				row = append(row, "")
			}
		}
		row = append(row, strconv.FormatInt(d.Volume, 10))
		rows[i] = row
	}
	return rows
}

func (a *analyticsOutput) JSONData() any { return a }

func (a *analyticsOutput) TemplateData() any { return a.Data }

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
	lf := &listFlags{}

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query DNS analytics",
		RunE: func(cmd *cobra.Command, args []string) error {
			effectiveGroupings, canonicalGroupings, err := normalizeGroupings(groupings)
			if err != nil {
				return err
			}

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
			if canonicalGroupings != "" {
				opts.Groupings = &canonicalGroupings
			}
			if sort != "" {
				opts.Sort = &sort
			}

			return runList(cmd, f, lf, "analytics rows",
				func(page, perPage int) ([]dnsimple.DnsAnalytics, *dnsimple.Pagination, error) {
					if page > 0 {
						opts.Page = &page
					}
					if perPage > 0 {
						opts.PerPage = &perPage
					}
					resp, err := c.DnsAnalytics.Query(context.Background(), accountID, opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				},
				func(items []dnsimple.DnsAnalytics, pg *dnsimple.Pagination) *analyticsOutput {
					return &analyticsOutput{Data: items, Pagination: pg, Groupings: effectiveGroupings}
				})
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO8601)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO8601)")
	cmd.Flags().StringVar(&groupings, "groupings", "", "Group by (comma-separated). Supported: zone_name, date")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	lf.register(cmd)

	return cmd
}

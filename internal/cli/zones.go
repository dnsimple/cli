package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/pagination"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

// zoneList adapts []Zone for output.
type zoneList struct {
	Data       []dnsimple.Zone      `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (z *zoneList) TableHeaders() []string {
	return []string{"ID", "NAME", "ACTIVE", "REVERSE"}
}

func (z *zoneList) TableRows() [][]string {
	rows := make([][]string, len(z.Data))
	for i, zone := range z.Data {
		rows[i] = []string{
			strconv.FormatInt(zone.ID, 10),
			zone.Name,
			strconv.FormatBool(zone.Active),
			strconv.FormatBool(zone.Reverse),
		}
	}
	return rows
}

func (z *zoneList) JSONData() any {
	return z
}

func (z *zoneList) TemplateData() any {
	return z.Data
}

// zoneItem adapts a single Zone for output.
type zoneItem struct {
	Data *dnsimple.Zone `json:"data"`
}

func (z *zoneItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (z *zoneItem) TableRows() [][]string {
	zone := z.Data
	return [][]string{
		{"ID", strconv.FormatInt(zone.ID, 10)},
		{"Name", zone.Name},
		{"Active", strconv.FormatBool(zone.Active)},
		{"Reverse", strconv.FormatBool(zone.Reverse)},
		{"Secondary", strconv.FormatBool(zone.Secondary)},
		{"Created At", zone.CreatedAt},
		{"Updated At", zone.UpdatedAt},
	}
}

func (z *zoneItem) JSONData() any {
	return z
}

func (z *zoneItem) TemplateData() any {
	return z.Data
}

func newZonesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zones",
		Short:   "Manage DNS zones",
		Aliases: []string{"zone"},
	}

	cmd.AddCommand(newZonesListCmd(f))
	cmd.AddCommand(newZonesGetCmd(f))
	cmd.AddCommand(newZonesRecordsCmd(f))
	cmd.AddCommand(newZonesFileCmd(f))
	cmd.AddCommand(newZonesActivateCmd(f))
	cmd.AddCommand(newZonesDeactivateCmd(f))
	cmd.AddCommand(newZonesCheckDistributionCmd(f))

	return cmd
}

func newZonesListCmd(f *cmdutil.Factory) *cobra.Command {
	var nameLike string
	var page, perPage int
	var sort string
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List zones",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			opts := &dnsimple.ZoneListOptions{}
			if nameLike != "" {
				opts.NameLike = &nameLike
			}
			if sort != "" {
				opts.Sort = &sort
			}

			if all {
				items, err := pagination.All(func(p int) ([]dnsimple.Zone, *dnsimple.Pagination, error) {
					opts.Page = &p
					resp, err := c.Zones.ListZones(context.Background(), accountID, opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				})
				if err != nil {
					return err
				}
				return f.Printer(cmd).Print(&zoneList{Data: items})
			}

			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}

			resp, err := c.Zones.ListZones(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&zoneList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().StringVar(&nameLike, "name-like", "", "Filter zones by name (partial match)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	return cmd
}

func newZonesGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <zone>",
		Short: "Get zone details",
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

			resp, err := c.Zones.GetZone(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&zoneItem{Data: resp.Data})
		},
	}
}

// formatRecordContent formats record content with type-specific display.
func formatRecordContent(r dnsimple.ZoneRecord) string {
	content := r.Content
	if r.Priority != 0 {
		content = fmt.Sprintf("%d %s", r.Priority, content)
	}
	return content
}

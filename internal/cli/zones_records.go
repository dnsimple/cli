package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/pagination"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// recordList adapts []ZoneRecord for output.
type recordList struct {
	Data       []dnsimple.ZoneRecord `json:"data"`
	Pagination *dnsimple.Pagination  `json:"pagination,omitempty"`
}

func (r *recordList) TableHeaders() []string {
	return []string{"ID", "TYPE", "NAME", "CONTENT", "TTL", "REGIONS"}
}

func (r *recordList) TableRows() [][]string {
	rows := make([][]string, len(r.Data))
	for i, rec := range r.Data {
		regions := ""
		if len(rec.Regions) > 0 {
			regions = strings.Join(rec.Regions, ",")
		}
		rows[i] = []string{
			strconv.FormatInt(rec.ID, 10),
			rec.Type,
			rec.Name,
			formatRecordContent(rec),
			strconv.Itoa(rec.TTL),
			regions,
		}
	}
	return rows
}

func (r *recordList) JSONData() any {
	return r
}

// recordItem adapts a single ZoneRecord for output.
type recordItem struct {
	Data *dnsimple.ZoneRecord `json:"data"`
}

func (r *recordItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (r *recordItem) TableRows() [][]string {
	rec := r.Data
	rows := [][]string{
		{"ID", strconv.FormatInt(rec.ID, 10)},
		{"Type", rec.Type},
		{"Name", rec.Name},
		{"Content", rec.Content},
		{"TTL", strconv.Itoa(rec.TTL)},
	}
	if rec.Priority != 0 {
		rows = append(rows, []string{"Priority", strconv.Itoa(rec.Priority)})
	}
	if len(rec.Regions) > 0 {
		rows = append(rows, []string{"Regions", strings.Join(rec.Regions, ", ")})
	}
	rows = append(rows,
		[]string{"System Record", strconv.FormatBool(rec.SystemRecord)},
		[]string{"Created At", rec.CreatedAt},
		[]string{"Updated At", rec.UpdatedAt},
	)
	return rows
}

func (r *recordItem) JSONData() any {
	return r
}

func newZonesRecordsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "records",
		Short:   "Manage zone records",
		Aliases: []string{"record"},
	}

	cmd.AddCommand(newRecordsListCmd(f))
	cmd.AddCommand(newRecordsGetCmd(f))
	cmd.AddCommand(newRecordsCreateCmd(f))
	cmd.AddCommand(newRecordsUpdateCmd(f))
	cmd.AddCommand(newRecordsDeleteCmd(f))
	cmd.AddCommand(newRecordsCheckDistributionCmd(f))

	return cmd
}

func newRecordsListCmd(f *cmdutil.Factory) *cobra.Command {
	var name, nameLike, recordType, sort string
	var page, perPage int
	var all bool

	cmd := &cobra.Command{
		Use:   "list <zone>",
		Short: "List zone records",
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

			opts := &dnsimple.ZoneRecordListOptions{}
			if name != "" {
				opts.Name = &name
			}
			if nameLike != "" {
				opts.NameLike = &nameLike
			}
			if recordType != "" {
				opts.Type = &recordType
			}
			if sort != "" {
				opts.Sort = &sort
			}

			if all {
				items, err := pagination.All(func(p int) ([]dnsimple.ZoneRecord, *dnsimple.Pagination, error) {
					opts.Page = &p
					resp, err := c.Zones.ListRecords(context.Background(), accountID, args[0], opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				})
				if err != nil {
					return err
				}
				return f.Printer(cmd).Print(&recordList{Data: items})
			}

			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}

			resp, err := c.Zones.ListRecords(context.Background(), accountID, args[0], opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&recordList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Filter by exact name")
	cmd.Flags().StringVar(&nameLike, "name-like", "", "Filter by name (partial match)")
	cmd.Flags().StringVar(&recordType, "type", "", "Filter by record type (A, AAAA, CNAME, MX, TXT, etc.)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	return cmd
}

func newRecordsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <zone> <record-id>",
		Short: "Get zone record details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			recordID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid record ID: %s", args[1])
			}

			resp, err := c.Zones.GetRecord(context.Background(), accountID, args[0], recordID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&recordItem{Data: resp.Data})
		},
	}
}

func newRecordsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var recordType, name, content string
	var ttl, priority int
	var regions []string

	cmd := &cobra.Command{
		Use:   "create <zone>",
		Short: "Create a zone record",
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

			attrs := dnsimple.ZoneRecordAttributes{
				Type:    recordType,
				Name:    &name,
				Content: content,
			}
			if ttl > 0 {
				attrs.TTL = ttl
			}
			if priority > 0 {
				attrs.Priority = priority
			}
			if len(regions) > 0 {
				attrs.Regions = regions
			}

			resp, err := c.Zones.CreateRecord(context.Background(), accountID, args[0], attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&recordItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&recordType, "type", "", "Record type (A, AAAA, CNAME, MX, TXT, etc.)")
	cmd.Flags().StringVar(&name, "name", "", "Record name (subdomain or empty for apex)")
	cmd.Flags().StringVar(&content, "content", "", "Record content")
	cmd.Flags().IntVar(&ttl, "ttl", 0, "Time to live in seconds")
	cmd.Flags().IntVar(&priority, "priority", 0, "Record priority (for MX, SRV)")
	cmd.Flags().StringSliceVar(&regions, "regions", nil, "Regions for regional records")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newRecordsUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	var name, content string
	var ttl, priority int
	var regions []string

	cmd := &cobra.Command{
		Use:   "update <zone> <record-id>",
		Short: "Update a zone record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			recordID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid record ID: %s", args[1])
			}

			attrs := dnsimple.ZoneRecordAttributes{}

			if cmd.Flags().Changed("name") {
				attrs.Name = &name
			}
			if cmd.Flags().Changed("content") {
				attrs.Content = content
			}
			if cmd.Flags().Changed("ttl") {
				attrs.TTL = ttl
			}
			if cmd.Flags().Changed("priority") {
				attrs.Priority = priority
			}
			if cmd.Flags().Changed("regions") {
				attrs.Regions = regions
			}

			resp, err := c.Zones.UpdateRecord(context.Background(), accountID, args[0], recordID, attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&recordItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Record name")
	cmd.Flags().StringVar(&content, "content", "", "Record content")
	cmd.Flags().IntVar(&ttl, "ttl", 0, "Time to live in seconds")
	cmd.Flags().IntVar(&priority, "priority", 0, "Record priority")
	cmd.Flags().StringSliceVar(&regions, "regions", nil, "Regions for regional records")

	return cmd
}

func newRecordsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <zone> <record-id>",
		Short: "Delete a zone record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			recordID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid record ID: %s", args[1])
			}

			_, err = c.Zones.DeleteRecord(context.Background(), accountID, args[0], recordID)
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Record %d deleted from zone %s\n", recordID, args[0])
			}
			return nil
		},
	}
}

package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/pagination"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// domainList adapts []Domain for output.
type domainList struct {
	Data       []dnsimple.Domain      `json:"data"`
	Pagination *dnsimple.Pagination   `json:"pagination,omitempty"`
}

func (d *domainList) TableHeaders() []string {
	return []string{"ID", "NAME", "STATE", "EXPIRES AT", "AUTO RENEW"}
}

func (d *domainList) TableRows() [][]string {
	rows := make([][]string, len(d.Data))
	for i, dom := range d.Data {
		rows[i] = []string{
			strconv.FormatInt(dom.ID, 10),
			dom.Name,
			dom.State,
			dom.ExpiresAt,
			strconv.FormatBool(dom.AutoRenew),
		}
	}
	return rows
}

func (d *domainList) JSONData() any {
	return d
}

// domainItem adapts a single Domain for output.
type domainItem struct {
	Data *dnsimple.Domain `json:"data"`
}

func (d *domainItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (d *domainItem) TableRows() [][]string {
	dom := d.Data
	return [][]string{
		{"ID", strconv.FormatInt(dom.ID, 10)},
		{"Name", dom.Name},
		{"State", dom.State},
		{"Auto Renew", strconv.FormatBool(dom.AutoRenew)},
		{"Private WHOIS", strconv.FormatBool(dom.PrivateWhois)},
		{"Expires At", dom.ExpiresAt},
		{"Created At", dom.CreatedAt},
		{"Updated At", dom.UpdatedAt},
	}
}

func (d *domainItem) JSONData() any {
	return d
}

func newDomainsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Short:   "Manage domains",
		Aliases: []string{"domain"},
	}

	cmd.AddCommand(newDomainsListCmd(f))
	cmd.AddCommand(newDomainsGetCmd(f))
	cmd.AddCommand(newDomainsCreateCmd(f))
	cmd.AddCommand(newDomainsDeleteCmd(f))
	cmd.AddCommand(newDomainsDnssecCmd(f))
	cmd.AddCommand(newDomainsDsRecordsCmd(f))
	cmd.AddCommand(newDomainsEmailForwardsCmd(f))
	cmd.AddCommand(newDomainsPushesCmd(f))

	return cmd
}

func newDomainsListCmd(f *cmdutil.Factory) *cobra.Command {
	var nameLike string
	var page, perPage int
	var sort string
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			opts := &dnsimple.DomainListOptions{}
			if nameLike != "" {
				opts.NameLike = &nameLike
			}
			if sort != "" {
				opts.Sort = &sort
			}

			if all {
				items, err := pagination.All(func(p int) ([]dnsimple.Domain, *dnsimple.Pagination, error) {
					opts.Page = &p
					resp, err := c.Domains.ListDomains(context.Background(), accountID, opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				})
				if err != nil {
					return err
				}
				return f.Printer().Print(&domainList{Data: items})
			}

			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}

			resp, err := c.Domains.ListDomains(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer().Print(&domainList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().StringVar(&nameLike, "name-like", "", "Filter domains by name (partial match)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order (e.g., name:asc, expiration:desc)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	return cmd
}

func newDomainsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <domain>",
		Short: "Get domain details",
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

			resp, err := c.Domains.GetDomain(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer().Print(&domainItem{Data: resp.Data})
		},
	}
}

func newDomainsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Add a domain to the account",
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

			resp, err := c.Domains.CreateDomain(context.Background(), accountID, dnsimple.Domain{Name: args[0]})
			if err != nil {
				return err
			}

			return f.Printer().Print(&domainItem{Data: resp.Data})
		},
	}
}

func newDomainsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <domain>",
		Short: "Delete a domain from the account",
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

			_, err = c.Domains.DeleteDomain(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Domain %s deleted\n", args[0])
			}
			return nil
		},
	}
}

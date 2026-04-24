package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/pagination"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

const confirmRegisteredDomainFlagHelp = "Acknowledge deletion of a registered domain, which downgrades it to hosted and loses registration metadata"

// domainList adapts []Domain for output.
type domainList struct {
	Data       []dnsimple.Domain    `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
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

func (d *domainList) TemplateData() any {
	return d.Data
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

func (d *domainItem) TemplateData() any {
	return d.Data
}

func newDomainsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage domains",
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
				return f.Printer(cmd).Print(&domainList{Data: items})
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

			return f.Printer(cmd).Print(&domainList{Data: resp.Data, Pagination: resp.Pagination})
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

			return f.Printer(cmd).Print(&domainItem{Data: resp.Data})
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

			return f.Printer(cmd).Print(&domainItem{Data: resp.Data})
		},
	}
}

func newDomainsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool
	var confirmRegisteredDomain bool

	cmd := &cobra.Command{
		Use:   "delete <domain>",
		Short: "Delete a domain from the account",
		Long: `Delete a domain from the account.

Hosted domains are deleted immediately. Registered domains are higher risk:
deleting them downgrades them to hosted and permanently removes registration
metadata such as expiry and auto-renew settings.

In scripts and CI, deleting a registered domain requires both --yes and
--confirm-registered-domain.`,
		Args: cobra.ExactArgs(1),
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

			if err := confirmDomainDeletion(cmd, resp.Data, yes, confirmRegisteredDomain); err != nil {
				return err
			}

			_, err = c.Domains.DeleteDomain(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Domain %s deleted\n", args[0])
			return nil
		},
	}

	addYesFlag(cmd, &yes)
	cmd.Flags().BoolVar(&confirmRegisteredDomain, "confirm-registered-domain", false, confirmRegisteredDomainFlagHelp)

	return cmd
}

func confirmDomainDeletion(cmd *cobra.Command, domain *dnsimple.Domain, yes bool, confirmRegisteredDomain bool) error {
	if domain == nil || domain.State != "registered" {
		name := ""
		if domain != nil {
			name = domain.Name
		}
		return confirmDestructiveAction(cmd, yes, fmt.Sprintf("Delete domain %s from the account?", name))
	}

	return confirmRegisteredDomainDeletionInput(
		cmd.InOrStdin(),
		cmd.ErrOrStderr(),
		isInteractiveInput(cmd.InOrStdin()),
		domain.Name,
		yes,
		confirmRegisteredDomain,
	)
}

func confirmRegisteredDomainDeletionInput(in io.Reader, errOut io.Writer, interactive bool, domain string, yes bool, confirmRegisteredDomain bool) error {
	if yes && confirmRegisteredDomain {
		return nil
	}
	if !interactive {
		return fmt.Errorf(
			"domain %s is registered; rerun with --yes --confirm-registered-domain to acknowledge the downgrade to hosted and loss of registration metadata",
			domain,
		)
	}

	fmt.Fprintf(
		errOut,
		"WARNING: %s is currently registered.\nThis will remove the registration from DNSimple, downgrade the domain to hosted,\nand permanently lose registration settings such as expiry and auto-renew.\n\nType the domain name to continue: ",
		domain,
	)

	answer, err := scanLine(in)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !strings.EqualFold(answer, domain) {
		return fmt.Errorf("confirmation declined")
	}
	return nil
}

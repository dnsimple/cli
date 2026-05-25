package cli

import (
	"context"
	"strconv"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type tldList struct {
	Data       []dnsimple.Tld       `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (t *tldList) TableHeaders() []string {
	return []string{"TLD", "TYPE", "REGISTRATION", "RENEWAL", "TRANSFER", "WHOIS PRIVACY"}
}

func (t *tldList) TableRows() [][]string {
	rows := make([][]string, len(t.Data))
	for i, tld := range t.Data {
		rows[i] = []string{
			tld.Tld,
			strconv.Itoa(tld.TldType),
			strconv.FormatBool(tld.RegistrationEnabled),
			strconv.FormatBool(tld.RenewalEnabled),
			strconv.FormatBool(tld.TransferEnabled),
			strconv.FormatBool(tld.WhoisPrivacy),
		}
	}
	return rows
}

func (t *tldList) JSONData() any { return t }

func (t *tldList) TemplateData() any { return t.Data }

type tldItemOutput struct {
	Data *dnsimple.Tld `json:"data"`
}

func (t *tldItemOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (t *tldItemOutput) TableRows() [][]string {
	tld := t.Data
	return [][]string{
		{"TLD", tld.Tld},
		{"Type", strconv.Itoa(tld.TldType)},
		{"Registration Enabled", strconv.FormatBool(tld.RegistrationEnabled)},
		{"Renewal Enabled", strconv.FormatBool(tld.RenewalEnabled)},
		{"Transfer Enabled", strconv.FormatBool(tld.TransferEnabled)},
		{"WHOIS Privacy", strconv.FormatBool(tld.WhoisPrivacy)},
		{"Auto Renew Only", strconv.FormatBool(tld.AutoRenewOnly)},
		{"Min Registration", strconv.Itoa(tld.MinimumRegistration)},
		{"DNSSEC Interface", tld.DnssecInterfaceType},
	}
}

func (t *tldItemOutput) JSONData() any { return t }

func (t *tldItemOutput) TemplateData() any { return t.Data }

type tldExtendedAttributesList struct {
	Data []dnsimple.TldExtendedAttribute `json:"data"`
}

func (t *tldExtendedAttributesList) TableHeaders() []string {
	return []string{"NAME", "DESCRIPTION", "REQUIRED"}
}

func (t *tldExtendedAttributesList) TableRows() [][]string {
	rows := make([][]string, len(t.Data))
	for i, attr := range t.Data {
		rows[i] = []string{
			attr.Name,
			attr.Description,
			strconv.FormatBool(attr.Required),
		}
	}
	return rows
}

func (t *tldExtendedAttributesList) JSONData() any { return t }

func (t *tldExtendedAttributesList) TemplateData() any { return t.Data }

func newTldsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tlds",
		Short: "List and inspect TLDs",
	}

	cmd.AddCommand(newTldsListCmd(f))
	cmd.AddCommand(newTldsGetCmd(f))
	cmd.AddCommand(newTldsExtendedAttributesCmd(f))

	return cmd
}

func newTldsListCmd(f *cmdutil.Factory) *cobra.Command {
	var page, perPage int
	var sort string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List supported TLDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			opts := &dnsimple.ListOptions{}
			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}
			if sort != "" {
				opts.Sort = &sort
			}

			resp, err := c.Tlds.ListTlds(context.Background(), opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).PrintList(&tldList{Data: resp.Data, Pagination: resp.Pagination}, pageHint(cmd, resp.Pagination, len(resp.Data), "TLDs"))
		},
	}

	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")

	return cmd
}

func newTldsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tld>",
		Short: "Get TLD details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Tlds.GetTld(context.Background(), args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&tldItemOutput{Data: resp.Data})
		},
	}
}

func newTldsExtendedAttributesCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "extended-attributes <tld>",
		Short: "Get extended attributes for a TLD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Tlds.GetTldExtendedAttributes(context.Background(), args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&tldExtendedAttributesList{Data: resp.Data})
		},
	}
}

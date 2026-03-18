package cli

import (
	"context"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// domainResearchStatusOutput adapts DomainResearchStatus for output.
type domainResearchStatusOutput struct {
	Data *dnsimple.DomainResearchStatus `json:"data"`
}

func (d *domainResearchStatusOutput) TableHeaders() []string {
	return []string{"DOMAIN", "AVAILABILITY", "ERRORS"}
}

func (d *domainResearchStatusOutput) TableRows() [][]string {
	return [][]string{{
		d.Data.Domain,
		d.Data.Availability,
		strings.Join(d.Data.Errors, ", "),
	}}
}

func (d *domainResearchStatusOutput) JSONData() any { return d }

func newResearchCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "research",
		Short: "Domain research operations",
	}

	cmd.AddCommand(newResearchStatusCmd(f))

	return cmd
}

func newResearchStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status <domain>",
		Short: "Check domain availability",
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

			resp, err := c.Domains.GetDomainResearchStatus(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainResearchStatusOutput{Data: resp.Data})
		},
	}
}

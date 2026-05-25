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

type serviceList struct {
	Data       []dnsimple.Service   `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (s *serviceList) TableHeaders() []string {
	return []string{"ID", "SID", "NAME", "DESCRIPTION"}
}

func (s *serviceList) TableRows() [][]string {
	rows := make([][]string, len(s.Data))
	for i, svc := range s.Data {
		rows[i] = []string{
			strconv.FormatInt(svc.ID, 10),
			svc.SID,
			svc.Name,
			svc.Description,
		}
	}
	return rows
}

func (s *serviceList) JSONData() any { return s }

func (s *serviceList) TemplateData() any { return s.Data }

type serviceItemOutput struct {
	Data *dnsimple.Service `json:"data"`
}

func (s *serviceItemOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (s *serviceItemOutput) TableRows() [][]string {
	svc := s.Data
	return [][]string{
		{"ID", strconv.FormatInt(svc.ID, 10)},
		{"SID", svc.SID},
		{"Name", svc.Name},
		{"Description", svc.Description},
		{"Setup Description", svc.SetupDescription},
		{"Requires Setup", strconv.FormatBool(svc.RequiresSetup)},
	}
}

func (s *serviceItemOutput) JSONData() any { return s }

func (s *serviceItemOutput) TemplateData() any { return s.Data }

func newServicesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage one-click services",
	}

	cmd.AddCommand(newServicesListCmd(f))
	cmd.AddCommand(newServicesGetCmd(f))
	cmd.AddCommand(newServicesAppliedCmd(f))
	cmd.AddCommand(newServicesApplyCmd(f))
	cmd.AddCommand(newServicesUnapplyCmd(f))

	return cmd
}

func newServicesListCmd(f *cmdutil.Factory) *cobra.Command {
	lf := &listFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available one-click services",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			opts := &dnsimple.ListOptions{}

			return runList(cmd, f, lf, "services",
				func(page, perPage int) ([]dnsimple.Service, *dnsimple.Pagination, error) {
					if page > 0 {
						opts.Page = &page
					}
					if perPage > 0 {
						opts.PerPage = &perPage
					}
					resp, err := c.Services.ListServices(context.Background(), opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				},
				func(items []dnsimple.Service, pg *dnsimple.Pagination) *serviceList {
					return &serviceList{Data: items, Pagination: pg}
				})
		},
	}

	lf.register(cmd)

	return cmd
}

func newServicesGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <service>",
		Short: "Get service details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Services.GetService(context.Background(), args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&serviceItemOutput{Data: resp.Data})
		},
	}
}

func newServicesAppliedCmd(f *cmdutil.Factory) *cobra.Command {
	lf := &listFlags{}

	cmd := &cobra.Command{
		Use:   "applied <domain>",
		Short: "List services applied to a domain",
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

			opts := &dnsimple.ListOptions{}

			return runList(cmd, f, lf, "applied services",
				func(page, perPage int) ([]dnsimple.Service, *dnsimple.Pagination, error) {
					if page > 0 {
						opts.Page = &page
					}
					if perPage > 0 {
						opts.PerPage = &perPage
					}
					resp, err := c.Services.AppliedServices(context.Background(), accountID, args[0], opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				},
				func(items []dnsimple.Service, pg *dnsimple.Pagination) *serviceList {
					return &serviceList{Data: items, Pagination: pg}
				})
		},
	}

	lf.register(cmd)

	return cmd
}

func newServicesApplyCmd(f *cmdutil.Factory) *cobra.Command {
	var settings []string

	cmd := &cobra.Command{
		Use:   "apply <domain> <service>",
		Short: "Apply a service to a domain",
		Example: `  dnsimple services apply example.com github-pages
  dnsimple services apply example.com heroku --settings app=my-heroku-app`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			service := args[1]

			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			s := dnsimple.DomainServiceSettings{Settings: make(map[string]string)}
			for _, kv := range settings {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					s.Settings[parts[0]] = parts[1]
				}
			}

			_, err = c.Services.ApplyService(context.Background(), accountID, service, domain, s)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Service %s applied to %s\n", service, domain)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&settings, "settings", nil, "Service settings (key=value)")

	return cmd
}

func newServicesUnapplyCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "unapply <domain> <service>",
		Short: "Remove a service from a domain",
		Example: `  dnsimple services unapply example.com github-pages
  dnsimple services unapply example.com github-pages --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			service := args[1]

			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Remove service %s from %s?", service, domain)); err != nil {
				return err
			}

			_, err = c.Services.UnapplyService(context.Background(), accountID, service, domain)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Service %s removed from %s\n", service, domain)
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

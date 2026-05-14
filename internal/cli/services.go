package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
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
	return &cobra.Command{
		Use:   "list",
		Short: "List available one-click services",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Services.ListServices(context.Background(), nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&serviceList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}
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
	return &cobra.Command{
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

			resp, err := c.Services.AppliedServices(context.Background(), accountID, args[0], nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&serviceList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}
}

func newServicesApplyCmd(f *cmdutil.Factory) *cobra.Command {
	var settings []string

	cmd := &cobra.Command{
		Use:   "apply <service> <domain>",
		Short: "Apply a service to a domain",
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

			s := dnsimple.DomainServiceSettings{Settings: make(map[string]string)}
			for _, kv := range settings {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					s.Settings[parts[0]] = parts[1]
				}
			}

			_, err = c.Services.ApplyService(context.Background(), accountID, args[0], args[1], s)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Service %s applied to %s\n", args[0], args[1])
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&settings, "settings", nil, "Service settings (key=value)")

	return cmd
}

func newServicesUnapplyCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "unapply <service> <domain>",
		Short: "Remove a service from a domain",
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

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Remove service %s from %s?", args[0], args[1])); err != nil {
				return err
			}

			_, err = c.Services.UnapplyService(context.Background(), accountID, args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Service %s removed from %s\n", args[0], args[1])
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

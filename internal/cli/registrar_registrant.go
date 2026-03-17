package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type registrantChangeList struct {
	Data       []dnsimple.RegistrantChange `json:"data"`
	Pagination *dnsimple.Pagination        `json:"pagination,omitempty"`
}

func (r *registrantChangeList) TableHeaders() []string {
	return []string{"ID", "DOMAIN ID", "CONTACT ID", "STATE"}
}

func (r *registrantChangeList) TableRows() [][]string {
	rows := make([][]string, len(r.Data))
	for i, rc := range r.Data {
		rows[i] = []string{
			strconv.Itoa(rc.Id),
			strconv.Itoa(rc.DomainId),
			strconv.Itoa(rc.ContactId),
			rc.State,
		}
	}
	return rows
}

func (r *registrantChangeList) JSONData() any { return r }

type registrantChangeOutput struct {
	Data *dnsimple.RegistrantChange `json:"data"`
}

func (r *registrantChangeOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (r *registrantChangeOutput) TableRows() [][]string {
	rc := r.Data
	return [][]string{
		{"ID", strconv.Itoa(rc.Id)},
		{"Domain ID", strconv.Itoa(rc.DomainId)},
		{"Contact ID", strconv.Itoa(rc.ContactId)},
		{"State", rc.State},
		{"Registry Owner Change", strconv.FormatBool(rc.RegistryOwnerChange)},
		{"Created At", rc.CreatedAt},
		{"Updated At", rc.UpdatedAt},
	}
}

func (r *registrantChangeOutput) JSONData() any { return r }

func newRegistrarRegistrantChangeCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registrant-change",
		Short: "Manage registrant changes",
	}

	cmd.AddCommand(newRegistrantChangeListCmd(f))
	cmd.AddCommand(newRegistrantChangeGetCmd(f))
	cmd.AddCommand(newRegistrantChangeCreateCmd(f))
	cmd.AddCommand(newRegistrantChangeDeleteCmd(f))

	return cmd
}

func newRegistrantChangeListCmd(f *cmdutil.Factory) *cobra.Command {
	var state, domainID, contactID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registrant changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			opts := &dnsimple.RegistrantChangeListOptions{}
			if state != "" {
				opts.State = &state
			}
			if domainID != "" {
				opts.DomainId = &domainID
			}
			if contactID != "" {
				opts.ContactId = &contactID
			}

			resp, err := c.Registrar.ListRegistrantChange(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&registrantChangeList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
	cmd.Flags().StringVar(&domainID, "domain-id", "", "Filter by domain ID")
	cmd.Flags().StringVar(&contactID, "contact-id", "", "Filter by contact ID")

	return cmd
}

func newRegistrantChangeGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <change-id>",
		Short: "Get registrant change details",
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

			changeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid change ID: %s", args[0])
			}

			resp, err := c.Registrar.GetRegistrantChange(context.Background(), accountID, changeID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&registrantChangeOutput{Data: resp.Data})
		},
	}
}

func newRegistrantChangeCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var domainID, contactID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a registrant change",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			input := &dnsimple.CreateRegistrantChangeInput{
				DomainId:  domainID,
				ContactId: contactID,
			}

			resp, err := c.Registrar.CreateRegistrantChange(context.Background(), accountID, input)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&registrantChangeOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&domainID, "domain-id", "", "Domain ID")
	cmd.Flags().StringVar(&contactID, "contact-id", "", "Contact ID")
	_ = cmd.MarkFlagRequired("domain-id")
	_ = cmd.MarkFlagRequired("contact-id")

	return cmd
}

func newRegistrantChangeDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <change-id>",
		Short: "Cancel a registrant change",
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

			changeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid change ID: %s", args[0])
			}

			_, err = c.Registrar.DeleteRegistrantChange(context.Background(), accountID, changeID)
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Registrant change %d cancelled\n", changeID)
			}
			return nil
		},
	}
}

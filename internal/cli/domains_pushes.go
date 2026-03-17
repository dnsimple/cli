package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// pushList adapts []DomainPush for output.
type pushList struct {
	Data       []dnsimple.DomainPush `json:"data"`
	Pagination *dnsimple.Pagination  `json:"pagination,omitempty"`
}

func (p *pushList) TableHeaders() []string {
	return []string{"ID", "DOMAIN ID", "ACCOUNT ID", "CREATED AT"}
}

func (p *pushList) TableRows() [][]string {
	rows := make([][]string, len(p.Data))
	for i, push := range p.Data {
		rows[i] = []string{
			strconv.FormatInt(push.ID, 10),
			strconv.FormatInt(push.DomainID, 10),
			strconv.FormatInt(push.AccountID, 10),
			push.CreatedAt,
		}
	}
	return rows
}

func (p *pushList) JSONData() any { return p }

// pushItem adapts a single DomainPush for output.
type pushItem struct {
	Data *dnsimple.DomainPush `json:"data"`
}

func (p *pushItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (p *pushItem) TableRows() [][]string {
	push := p.Data
	return [][]string{
		{"ID", strconv.FormatInt(push.ID, 10)},
		{"Domain ID", strconv.FormatInt(push.DomainID, 10)},
		{"Account ID", strconv.FormatInt(push.AccountID, 10)},
		{"Contact ID", strconv.FormatInt(push.ContactID, 10)},
		{"Created At", push.CreatedAt},
		{"Accepted At", push.AcceptedAt},
	}
}

func (p *pushItem) JSONData() any { return p }

func newDomainsPushesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pushes",
		Short: "Manage domain pushes",
	}

	cmd.AddCommand(newPushesListCmd(f))
	cmd.AddCommand(newPushesInitiateCmd(f))
	cmd.AddCommand(newPushesAcceptCmd(f))
	cmd.AddCommand(newPushesRejectCmd(f))

	return cmd
}

func newPushesListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pending domain pushes",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Domains.ListPushes(context.Background(), accountID, nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&pushList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}
}

func newPushesInitiateCmd(f *cmdutil.Factory) *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use:   "initiate <domain>",
		Short: "Push a domain to another account",
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

			attrs := dnsimple.DomainPushAttributes{
				NewAccountEmail: email,
			}

			resp, err := c.Domains.InitiatePush(context.Background(), accountID, args[0], attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&pushItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email of the target account")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func newPushesAcceptCmd(f *cmdutil.Factory) *cobra.Command {
	var contactID int64

	cmd := &cobra.Command{
		Use:   "accept <push-id>",
		Short: "Accept a domain push",
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

			pushID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid push ID: %s", args[0])
			}

			attrs := dnsimple.DomainPushAttributes{
				ContactID: contactID,
			}

			_, err = c.Domains.AcceptPush(context.Background(), accountID, pushID, attrs)
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Push %d accepted\n", pushID)
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&contactID, "contact-id", 0, "Contact ID for the domain")
	_ = cmd.MarkFlagRequired("contact-id")

	return cmd
}

func newPushesRejectCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "reject <push-id>",
		Short: "Reject a domain push",
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

			pushID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid push ID: %s", args[0])
			}

			_, err = c.Domains.RejectPush(context.Background(), accountID, pushID)
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Push %d rejected\n", pushID)
			}
			return nil
		},
	}
}

package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// emailForwardList adapts []EmailForward for output.
type emailForwardList struct {
	Data       []dnsimple.EmailForward `json:"data"`
	Pagination *dnsimple.Pagination    `json:"pagination,omitempty"`
}

func (e *emailForwardList) TableHeaders() []string {
	return []string{"ID", "FROM", "TO", "ACTIVE"}
}

func (e *emailForwardList) TableRows() [][]string {
	rows := make([][]string, len(e.Data))
	for i, ef := range e.Data {
		rows[i] = []string{
			strconv.FormatInt(ef.ID, 10),
			ef.AliasEmail,
			ef.DestinationEmail,
			strconv.FormatBool(ef.Active),
		}
	}
	return rows
}

func (e *emailForwardList) JSONData() any { return e }

func (e *emailForwardList) TemplateData() any { return e.Data }

// emailForwardItem adapts a single EmailForward for output.
type emailForwardItem struct {
	Data *dnsimple.EmailForward `json:"data"`
}

func (e *emailForwardItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (e *emailForwardItem) TableRows() [][]string {
	ef := e.Data
	return [][]string{
		{"ID", strconv.FormatInt(ef.ID, 10)},
		{"From", ef.AliasEmail},
		{"To", ef.DestinationEmail},
		{"Active", strconv.FormatBool(ef.Active)},
		{"Created At", ef.CreatedAt},
		{"Updated At", ef.UpdatedAt},
	}
}

func (e *emailForwardItem) JSONData() any { return e }

func (e *emailForwardItem) TemplateData() any { return e.Data }

func newDomainsEmailForwardsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email-forwards",
		Short: "Manage email forwards",
	}

	cmd.AddCommand(newEmailForwardsListCmd(f))
	cmd.AddCommand(newEmailForwardsGetCmd(f))
	cmd.AddCommand(newEmailForwardsCreateCmd(f))
	cmd.AddCommand(newEmailForwardsDeleteCmd(f))

	return cmd
}

func newEmailForwardsListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list <domain>",
		Short: "List email forwards",
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

			resp, err := c.Domains.ListEmailForwards(context.Background(), accountID, args[0], nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).PrintList(&emailForwardList{Data: resp.Data, Pagination: resp.Pagination}, pageHint(cmd, resp.Pagination, len(resp.Data), "email forwards"))
		},
	}
}

func newEmailForwardsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <domain> <forward-id>",
		Short: "Get email forward details",
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

			forwardID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid forward ID: %s", args[1])
			}

			resp, err := c.Domains.GetEmailForward(context.Background(), accountID, args[0], forwardID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&emailForwardItem{Data: resp.Data})
		},
	}
}

func newEmailForwardsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var from, to string

	cmd := &cobra.Command{
		Use:   "create <domain>",
		Short: "Create an email forward",
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

			attrs := dnsimple.EmailForward{
				AliasName:        from,
				DestinationEmail: to,
			}

			resp, err := c.Domains.CreateEmailForward(context.Background(), accountID, args[0], attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&emailForwardItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Alias name (e.g., info)")
	cmd.Flags().StringVar(&to, "to", "", "Destination email address")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func newEmailForwardsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <domain> <forward-id>",
		Short: "Delete an email forward",
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

			forwardID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid forward ID: %s", args[1])
			}

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Delete email forward %d from %s?", forwardID, args[0])); err != nil {
				return err
			}

			_, err = c.Domains.DeleteEmailForward(context.Background(), accountID, args[0], forwardID)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Email forward %d deleted from %s\n", forwardID, args[0])
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

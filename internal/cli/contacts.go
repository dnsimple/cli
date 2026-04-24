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

// contactList adapts []Contact for output.
type contactList struct {
	Data       []dnsimple.Contact   `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (c *contactList) TableHeaders() []string {
	return []string{"ID", "LABEL", "NAME", "EMAIL", "PHONE"}
}

func (c *contactList) TableRows() [][]string {
	rows := make([][]string, len(c.Data))
	for i, ct := range c.Data {
		rows[i] = []string{
			strconv.FormatInt(ct.ID, 10),
			ct.Label,
			ct.FirstName + " " + ct.LastName,
			ct.Email,
			ct.Phone,
		}
	}
	return rows
}

func (c *contactList) JSONData() any { return c }

func (c *contactList) TemplateData() any { return c.Data }

// contactItem adapts a single Contact for output.
type contactItem struct {
	Data *dnsimple.Contact `json:"data"`
}

func (c *contactItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (c *contactItem) TableRows() [][]string {
	ct := c.Data
	rows := [][]string{
		{"ID", strconv.FormatInt(ct.ID, 10)},
		{"Label", ct.Label},
		{"First Name", ct.FirstName},
		{"Last Name", ct.LastName},
		{"Email", ct.Email},
		{"Phone", ct.Phone},
		{"Organization", ct.Organization},
		{"Address", ct.Address1},
	}
	if ct.Address2 != "" {
		rows = append(rows, []string{"Address 2", ct.Address2})
	}
	rows = append(rows,
		[]string{"City", ct.City},
		[]string{"State/Province", ct.StateProvince},
		[]string{"Postal Code", ct.PostalCode},
		[]string{"Country", ct.Country},
		[]string{"Created At", ct.CreatedAt},
		[]string{"Updated At", ct.UpdatedAt},
	)
	return rows
}

func (c *contactItem) JSONData() any { return c }

func (c *contactItem) TemplateData() any { return c.Data }

func newContactsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage contacts",
	}

	cmd.AddCommand(newContactsListCmd(f))
	cmd.AddCommand(newContactsGetCmd(f))
	cmd.AddCommand(newContactsCreateCmd(f))
	cmd.AddCommand(newContactsUpdateCmd(f))
	cmd.AddCommand(newContactsDeleteCmd(f))

	return cmd
}

func newContactsListCmd(f *cmdutil.Factory) *cobra.Command {
	var page, perPage int
	var sort string
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
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
			if sort != "" {
				opts.Sort = &sort
			}

			if all {
				items, err := pagination.All(func(p int) ([]dnsimple.Contact, *dnsimple.Pagination, error) {
					opts.Page = &p
					resp, err := c.Contacts.ListContacts(context.Background(), accountID, opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				})
				if err != nil {
					return err
				}
				return f.Printer(cmd).Print(&contactList{Data: items})
			}

			if page > 0 {
				opts.Page = &page
			}
			if perPage > 0 {
				opts.PerPage = &perPage
			}

			resp, err := c.Contacts.ListContacts(context.Background(), accountID, opts)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&contactList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Number of items per page")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	return cmd
}

func newContactsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <contact-id>",
		Short: "Get contact details",
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

			contactID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid contact ID: %s", args[0])
			}

			resp, err := c.Contacts.GetContact(context.Background(), accountID, contactID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&contactItem{Data: resp.Data})
		},
	}
}

func newContactsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var contact dnsimple.Contact

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Contacts.CreateContact(context.Background(), accountID, contact)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&contactItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&contact.Label, "label", "", "Contact label")
	cmd.Flags().StringVar(&contact.FirstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&contact.LastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&contact.Email, "email", "", "Email address")
	cmd.Flags().StringVar(&contact.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&contact.Organization, "organization", "", "Organization name")
	cmd.Flags().StringVar(&contact.JobTitle, "job-title", "", "Job title")
	cmd.Flags().StringVar(&contact.Address1, "address1", "", "Address line 1")
	cmd.Flags().StringVar(&contact.Address2, "address2", "", "Address line 2")
	cmd.Flags().StringVar(&contact.City, "city", "", "City")
	cmd.Flags().StringVar(&contact.StateProvince, "state-province", "", "State or province")
	cmd.Flags().StringVar(&contact.PostalCode, "postal-code", "", "Postal code")
	cmd.Flags().StringVar(&contact.Country, "country", "", "Country code (e.g., US, IT)")
	cmd.Flags().StringVar(&contact.Fax, "fax", "", "Fax number")

	_ = cmd.MarkFlagRequired("first-name")
	_ = cmd.MarkFlagRequired("last-name")
	_ = cmd.MarkFlagRequired("address1")
	_ = cmd.MarkFlagRequired("city")
	_ = cmd.MarkFlagRequired("state-province")
	_ = cmd.MarkFlagRequired("postal-code")
	_ = cmd.MarkFlagRequired("country")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("phone")

	return cmd
}

func newContactsUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	var contact dnsimple.Contact

	cmd := &cobra.Command{
		Use:   "update <contact-id>",
		Short: "Update a contact",
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

			contactID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid contact ID: %s", args[0])
			}

			resp, err := c.Contacts.UpdateContact(context.Background(), accountID, contactID, contact)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&contactItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&contact.Label, "label", "", "Contact label")
	cmd.Flags().StringVar(&contact.FirstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&contact.LastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&contact.Email, "email", "", "Email address")
	cmd.Flags().StringVar(&contact.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&contact.Organization, "organization", "", "Organization name")
	cmd.Flags().StringVar(&contact.JobTitle, "job-title", "", "Job title")
	cmd.Flags().StringVar(&contact.Address1, "address1", "", "Address line 1")
	cmd.Flags().StringVar(&contact.Address2, "address2", "", "Address line 2")
	cmd.Flags().StringVar(&contact.City, "city", "", "City")
	cmd.Flags().StringVar(&contact.StateProvince, "state-province", "", "State or province")
	cmd.Flags().StringVar(&contact.PostalCode, "postal-code", "", "Postal code")
	cmd.Flags().StringVar(&contact.Country, "country", "", "Country code (e.g., US, IT)")
	cmd.Flags().StringVar(&contact.Fax, "fax", "", "Fax number")

	return cmd
}

func newContactsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <contact-id>",
		Short: "Delete a contact",
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

			contactID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid contact ID: %s", args[0])
			}

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Delete contact %d?", contactID)); err != nil {
				return err
			}

			_, err = c.Contacts.DeleteContact(context.Background(), accountID, contactID)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Contact %d deleted\n", contactID)
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

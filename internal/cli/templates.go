package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type templateList struct {
	Data       []dnsimple.Template  `json:"data"`
	Pagination *dnsimple.Pagination `json:"pagination,omitempty"`
}

func (t *templateList) TableHeaders() []string {
	return []string{"ID", "SID", "NAME", "DESCRIPTION"}
}

func (t *templateList) TableRows() [][]string {
	rows := make([][]string, len(t.Data))
	for i, tmpl := range t.Data {
		rows[i] = []string{
			strconv.FormatInt(tmpl.ID, 10),
			tmpl.SID,
			tmpl.Name,
			tmpl.Description,
		}
	}
	return rows
}

func (t *templateList) JSONData() any { return t }

type templateItemOutput struct {
	Data *dnsimple.Template `json:"data"`
}

func (t *templateItemOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (t *templateItemOutput) TableRows() [][]string {
	tmpl := t.Data
	return [][]string{
		{"ID", strconv.FormatInt(tmpl.ID, 10)},
		{"SID", tmpl.SID},
		{"Name", tmpl.Name},
		{"Description", tmpl.Description},
		{"Created At", tmpl.CreatedAt},
		{"Updated At", tmpl.UpdatedAt},
	}
}

func (t *templateItemOutput) JSONData() any { return t }

func newTemplatesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "templates",
		Short:   "Manage DNS templates",
		Aliases: []string{"template"},
	}

	cmd.AddCommand(newTemplatesListCmd(f))
	cmd.AddCommand(newTemplatesGetCmd(f))
	cmd.AddCommand(newTemplatesCreateCmd(f))
	cmd.AddCommand(newTemplatesUpdateCmd(f))
	cmd.AddCommand(newTemplatesDeleteCmd(f))
	cmd.AddCommand(newTemplatesApplyCmd(f))
	cmd.AddCommand(newTemplatesRecordsCmd(f))

	return cmd
}

func newTemplatesListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Templates.ListTemplates(context.Background(), accountID, nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}
}

func newTemplatesGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <template>",
		Short: "Get template details",
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

			resp, err := c.Templates.GetTemplate(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateItemOutput{Data: resp.Data})
		},
	}
}

func newTemplatesCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var tmpl dnsimple.Template

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a template",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Templates.CreateTemplate(context.Background(), accountID, tmpl)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateItemOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&tmpl.Name, "name", "", "Template name")
	cmd.Flags().StringVar(&tmpl.SID, "sid", "", "Template short ID")
	cmd.Flags().StringVar(&tmpl.Description, "description", "", "Template description")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("sid")

	return cmd
}

func newTemplatesUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	var tmpl dnsimple.Template

	cmd := &cobra.Command{
		Use:   "update <template>",
		Short: "Update a template",
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

			resp, err := c.Templates.UpdateTemplate(context.Background(), accountID, args[0], tmpl)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateItemOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&tmpl.Name, "name", "", "Template name")
	cmd.Flags().StringVar(&tmpl.SID, "sid", "", "Template short ID")
	cmd.Flags().StringVar(&tmpl.Description, "description", "", "Template description")

	return cmd
}

func newTemplatesDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <template>",
		Short: "Delete a template",
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

			_, err = c.Templates.DeleteTemplate(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Template %s deleted\n", args[0])
			}
			return nil
		},
	}
}

func newTemplatesApplyCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "apply <template> <domain>",
		Short: "Apply a template to a domain",
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

			_, err = c.Templates.ApplyTemplate(context.Background(), accountID, args[0], args[1])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Template %s applied to %s\n", args[0], args[1])
			}
			return nil
		},
	}
}

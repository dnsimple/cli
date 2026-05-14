package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

type templateRecordList struct {
	Data       []dnsimple.TemplateRecord `json:"data"`
	Pagination *dnsimple.Pagination      `json:"pagination,omitempty"`
}

func (t *templateRecordList) TableHeaders() []string {
	return []string{"ID", "TYPE", "NAME", "CONTENT", "TTL"}
}

func (t *templateRecordList) TableRows() [][]string {
	rows := make([][]string, len(t.Data))
	for i, rec := range t.Data {
		rows[i] = []string{
			strconv.FormatInt(rec.ID, 10),
			rec.Type,
			rec.Name,
			rec.Content,
			strconv.Itoa(rec.TTL),
		}
	}
	return rows
}

func (t *templateRecordList) JSONData() any { return t }

func (t *templateRecordList) TemplateData() any { return t.Data }

type templateRecordItemOutput struct {
	Data *dnsimple.TemplateRecord `json:"data"`
}

func (t *templateRecordItemOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (t *templateRecordItemOutput) TableRows() [][]string {
	rec := t.Data
	rows := [][]string{
		{"ID", strconv.FormatInt(rec.ID, 10)},
		{"Type", rec.Type},
		{"Name", rec.Name},
		{"Content", rec.Content},
		{"TTL", strconv.Itoa(rec.TTL)},
	}
	if rec.Priority != 0 {
		rows = append(rows, []string{"Priority", strconv.Itoa(rec.Priority)})
	}
	return rows
}

func (t *templateRecordItemOutput) JSONData() any { return t }

func (t *templateRecordItemOutput) TemplateData() any { return t.Data }

func newTemplatesRecordsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "Manage template records",
	}

	cmd.AddCommand(newTemplateRecordsListCmd(f))
	cmd.AddCommand(newTemplateRecordsGetCmd(f))
	cmd.AddCommand(newTemplateRecordsCreateCmd(f))
	cmd.AddCommand(newTemplateRecordsDeleteCmd(f))

	return cmd
}

func newTemplateRecordsListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list <template>",
		Short: "List template records",
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

			resp, err := c.Templates.ListTemplateRecords(context.Background(), accountID, args[0], nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateRecordList{Data: resp.Data, Pagination: resp.Pagination})
		},
	}
}

func newTemplateRecordsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <template> <record-id>",
		Short: "Get template record details",
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

			recordID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid record ID: %s", args[1])
			}

			resp, err := c.Templates.GetTemplateRecord(context.Background(), accountID, args[0], recordID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateRecordItemOutput{Data: resp.Data})
		},
	}
}

func newTemplateRecordsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var rec dnsimple.TemplateRecord

	cmd := &cobra.Command{
		Use:   "create <template>",
		Short: "Create a template record",
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

			resp, err := c.Templates.CreateTemplateRecord(context.Background(), accountID, args[0], rec)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&templateRecordItemOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&rec.Type, "type", "", "Record type")
	cmd.Flags().StringVar(&rec.Name, "name", "", "Record name")
	cmd.Flags().StringVar(&rec.Content, "content", "", "Record content")
	cmd.Flags().IntVar(&rec.TTL, "ttl", 0, "Time to live")
	cmd.Flags().IntVar(&rec.Priority, "priority", 0, "Record priority")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newTemplateRecordsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <template> <record-id>",
		Short: "Delete a template record",
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

			recordID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid record ID: %s", args[1])
			}

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Delete template record %d from %s?", recordID, args[0])); err != nil {
				return err
			}

			_, err = c.Templates.DeleteTemplateRecord(context.Background(), accountID, args[0], recordID)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Template record %d deleted\n", recordID)
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

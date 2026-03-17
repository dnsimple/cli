package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

type webhookList struct {
	Data []dnsimple.Webhook `json:"data"`
}

func (w *webhookList) TableHeaders() []string {
	return []string{"ID", "URL"}
}

func (w *webhookList) TableRows() [][]string {
	rows := make([][]string, len(w.Data))
	for i, wh := range w.Data {
		rows[i] = []string{
			strconv.FormatInt(wh.ID, 10),
			wh.URL,
		}
	}
	return rows
}

func (w *webhookList) JSONData() any { return w }

type webhookItemOutput struct {
	Data *dnsimple.Webhook `json:"data"`
}

func (w *webhookItemOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (w *webhookItemOutput) TableRows() [][]string {
	return [][]string{
		{"ID", strconv.FormatInt(w.Data.ID, 10)},
		{"URL", w.Data.URL},
	}
}

func (w *webhookItemOutput) JSONData() any { return w }

func newWebhooksCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage webhooks",
	}

	cmd.AddCommand(newWebhooksListCmd(f))
	cmd.AddCommand(newWebhooksGetCmd(f))
	cmd.AddCommand(newWebhooksCreateCmd(f))
	cmd.AddCommand(newWebhooksDeleteCmd(f))

	return cmd
}

func newWebhooksListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Webhooks.ListWebhooks(context.Background(), accountID, nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&webhookList{Data: resp.Data})
		},
	}
}

func newWebhooksGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <webhook-id>",
		Short: "Get webhook details",
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

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[0])
			}

			resp, err := c.Webhooks.GetWebhook(context.Background(), accountID, id)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&webhookItemOutput{Data: resp.Data})
		},
	}
}

func newWebhooksCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var url string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resp, err := c.Webhooks.CreateWebhook(context.Background(), accountID, dnsimple.Webhook{URL: url})
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&webhookItemOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "Webhook URL")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newWebhooksDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <webhook-id>",
		Short: "Delete a webhook",
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

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[0])
			}

			_, err = c.Webhooks.DeleteWebhook(context.Background(), accountID, id)
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Webhook %d deleted\n", id)
			}
			return nil
		},
	}
}

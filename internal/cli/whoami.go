package cli

import (
	"context"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

type whoamiOutput struct {
	UserID       int64  `json:"user_id,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	AccountID    int64  `json:"account_id,omitempty"`
	AccountEmail string `json:"account_email,omitempty"`
}

func (w *whoamiOutput) TableHeaders() []string {
	return []string{"TYPE", "ID", "EMAIL"}
}

func (w *whoamiOutput) TableRows() [][]string {
	var rows [][]string
	if w.UserID != 0 {
		rows = append(rows, []string{"User", strconv.FormatInt(w.UserID, 10), w.UserEmail})
	}
	if w.AccountID != 0 {
		rows = append(rows, []string{"Account", strconv.FormatInt(w.AccountID, 10), w.AccountEmail})
	}
	return rows
}

func (w *whoamiOutput) JSONData() any {
	return w
}

func newWhoamiCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the currently authenticated identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			resp, err := c.Identity.Whoami(context.Background())
			if err != nil {
				return err
			}

			data := &whoamiOutput{}
			if resp.Data.User != nil {
				data.UserID = resp.Data.User.ID
				data.UserEmail = resp.Data.User.Email
			}
			if resp.Data.Account != nil {
				data.AccountID = resp.Data.Account.ID
				data.AccountEmail = resp.Data.Account.Email
			}

			return f.Printer(cmd).Print(data)
		},
	}
}

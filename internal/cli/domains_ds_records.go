package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// dsRecordList adapts []DelegationSignerRecord for output.
type dsRecordList struct {
	Data       []dnsimple.DelegationSignerRecord `json:"data"`
	Pagination *dnsimple.Pagination              `json:"pagination,omitempty"`
}

func (d *dsRecordList) TableHeaders() []string {
	return []string{"ID", "ALGORITHM", "DIGEST", "DIGEST TYPE", "KEYTAG"}
}

func (d *dsRecordList) TableRows() [][]string {
	rows := make([][]string, len(d.Data))
	for i, ds := range d.Data {
		rows[i] = []string{
			strconv.FormatInt(ds.ID, 10),
			ds.Algorithm,
			ds.Digest,
			ds.DigestType,
			ds.Keytag,
		}
	}
	return rows
}

func (d *dsRecordList) JSONData() any { return d }

func (d *dsRecordList) TemplateData() any { return d.Data }

// dsRecordItem adapts a single DelegationSignerRecord for output.
type dsRecordItem struct {
	Data *dnsimple.DelegationSignerRecord `json:"data"`
}

func (d *dsRecordItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (d *dsRecordItem) TableRows() [][]string {
	ds := d.Data
	return [][]string{
		{"ID", strconv.FormatInt(ds.ID, 10)},
		{"Algorithm", ds.Algorithm},
		{"Digest", ds.Digest},
		{"Digest Type", ds.DigestType},
		{"Keytag", ds.Keytag},
		{"Public Key", ds.PublicKey},
		{"Created At", ds.CreatedAt},
		{"Updated At", ds.UpdatedAt},
	}
}

func (d *dsRecordItem) JSONData() any { return d }

func (d *dsRecordItem) TemplateData() any { return d.Data }

func newDomainsDsRecordsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ds-records",
		Short: "Manage delegation signer (DS) records",
	}

	cmd.AddCommand(newDsRecordsListCmd(f))
	cmd.AddCommand(newDsRecordsGetCmd(f))
	cmd.AddCommand(newDsRecordsCreateCmd(f))
	cmd.AddCommand(newDsRecordsDeleteCmd(f))

	return cmd
}

func newDsRecordsListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list <domain>",
		Short: "List DS records",
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

			resp, err := c.Domains.ListDelegationSignerRecords(context.Background(), accountID, args[0], nil)
			if err != nil {
				return err
			}

			return f.Printer(cmd).PrintList(&dsRecordList{Data: resp.Data, Pagination: resp.Pagination}, pageHint(cmd, resp.Pagination, len(resp.Data), "DS records"))
		},
	}
}

func newDsRecordsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <domain> <ds-record-id>",
		Short: "Get DS record details",
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

			dsID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DS record ID: %s", args[1])
			}

			resp, err := c.Domains.GetDelegationSignerRecord(context.Background(), accountID, args[0], dsID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&dsRecordItem{Data: resp.Data})
		},
	}
}

func newDsRecordsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var ds dnsimple.DelegationSignerRecord

	cmd := &cobra.Command{
		Use:   "create <domain>",
		Short: "Create a DS record",
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

			resp, err := c.Domains.CreateDelegationSignerRecord(context.Background(), accountID, args[0], ds)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&dsRecordItem{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&ds.Algorithm, "algorithm", "", "DNSSEC algorithm")
	cmd.Flags().StringVar(&ds.Digest, "digest", "", "Digest value")
	cmd.Flags().StringVar(&ds.DigestType, "digest-type", "", "Digest type")
	cmd.Flags().StringVar(&ds.Keytag, "keytag", "", "Key tag")
	cmd.Flags().StringVar(&ds.PublicKey, "public-key", "", "Public key")
	_ = cmd.MarkFlagRequired("algorithm")

	return cmd
}

func newDsRecordsDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <domain> <ds-record-id>",
		Short: "Delete a DS record",
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

			dsID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DS record ID: %s", args[1])
			}

			if err := confirmDestructiveAction(cmd, yes, fmt.Sprintf("Delete DS record %d from %s?", dsID, args[0])); err != nil {
				return err
			}

			_, err = c.Domains.DeleteDelegationSignerRecord(context.Background(), accountID, args[0], dsID)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "DS record %d deleted from %s\n", dsID, args[0])
			return nil
		},
	}

	addYesFlag(cmd, &yes)

	return cmd
}

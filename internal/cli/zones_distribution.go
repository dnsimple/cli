package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// distributionOutput adapts ZoneDistribution for output.
type distributionOutput struct {
	Data *dnsimple.ZoneDistribution `json:"data"`
}

func (d *distributionOutput) TableHeaders() []string {
	return []string{"DISTRIBUTED"}
}

func (d *distributionOutput) TableRows() [][]string {
	return [][]string{{strconv.FormatBool(d.Data.Distributed)}}
}

func (d *distributionOutput) JSONData() any { return d }

// zoneFileOutput adapts ZoneFile for output.
type zoneFileOutput struct {
	Data *dnsimple.ZoneFile `json:"data"`
}

func (z *zoneFileOutput) TableHeaders() []string {
	return []string{}
}

func (z *zoneFileOutput) TableRows() [][]string {
	return nil
}

func (z *zoneFileOutput) JSONData() any { return z }

func newZonesFileCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "file <zone>",
		Short: "Download zone file",
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

			resp, err := c.Zones.GetZoneFile(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if f.Flags.JSON || f.Flags.Format != "" {
				return f.Printer(cmd).Print(&zoneFileOutput{Data: resp.Data})
			}

			// In table mode, just print the raw zone file content
			fmt.Fprint(cmd.OutOrStdout(), resp.Data.Zone)
			return nil
		},
	}
}

func newZonesActivateCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "activate <zone>",
		Short: "Activate DNS for a zone",
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

			_, err = c.Zones.ActivateZoneDns(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "DNS activated for zone %s\n", args[0])
			}
			return nil
		},
	}
}

func newZonesDeactivateCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "deactivate <zone>",
		Short: "Deactivate DNS for a zone",
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

			_, err = c.Zones.DeactivateZoneDns(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			if !f.Flags.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "DNS deactivated for zone %s\n", args[0])
			}
			return nil
		},
	}
}

func newZonesCheckDistributionCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "check-distribution <zone>",
		Short: "Check zone distribution status",
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

			resp, err := c.Zones.CheckZoneDistribution(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&distributionOutput{Data: resp.Data})
		},
	}
}

func newRecordsCheckDistributionCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "check-distribution <zone> <record-id>",
		Short: "Check record distribution status",
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

			resp, err := c.Zones.CheckZoneRecordDistribution(context.Background(), accountID, args[0], recordID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&distributionOutput{Data: resp.Data})
		},
	}
}

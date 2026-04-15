package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// domainCheckOutput adapts DomainCheck for output.
type domainCheckOutput struct {
	Data *dnsimple.DomainCheck `json:"data"`
}

func (d *domainCheckOutput) TableHeaders() []string {
	return []string{"DOMAIN", "AVAILABLE", "PREMIUM"}
}

func (d *domainCheckOutput) TableRows() [][]string {
	return [][]string{{
		d.Data.Domain,
		strconv.FormatBool(d.Data.Available),
		strconv.FormatBool(d.Data.Premium),
	}}
}

func (d *domainCheckOutput) JSONData() any { return d }

func (d *domainCheckOutput) TemplateData() any { return d.Data }

// domainPriceOutput adapts DomainPrice for output.
type domainPriceOutput struct {
	Data *dnsimple.DomainPrice `json:"data"`
}

func (d *domainPriceOutput) TableHeaders() []string {
	return []string{"DOMAIN", "PREMIUM", "REGISTRATION", "RENEWAL", "TRANSFER", "TRUSTEE"}
}

func (d *domainPriceOutput) TableRows() [][]string {
	trustee := "-"
	if d.Data.TrusteeServicePrice != nil {
		trustee = fmt.Sprintf("%.2f", *d.Data.TrusteeServicePrice)
	}
	return [][]string{{
		d.Data.Domain,
		strconv.FormatBool(d.Data.Premium),
		fmt.Sprintf("%.2f", d.Data.RegistrationPrice),
		fmt.Sprintf("%.2f", d.Data.RenewalPrice),
		fmt.Sprintf("%.2f", d.Data.TransferPrice),
		trustee,
	}}
}

func (d *domainPriceOutput) JSONData() any { return d }

func (d *domainPriceOutput) TemplateData() any { return d.Data }

// domainRegistrationOutput adapts DomainRegistration for output.
type domainRegistrationOutput struct {
	Data *dnsimple.DomainRegistration `json:"data"`
}

func (d *domainRegistrationOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (d *domainRegistrationOutput) TableRows() [][]string {
	r := d.Data
	return [][]string{
		{"ID", strconv.FormatInt(r.ID, 10)},
		{"Domain ID", strconv.FormatInt(r.DomainID, 10)},
		{"Registrant ID", strconv.FormatInt(r.RegistrantID, 10)},
		{"State", r.State},
		{"Auto Renew", strconv.FormatBool(r.AutoRenew)},
		{"WHOIS Privacy", strconv.FormatBool(r.WhoisPrivacy)},
		{"Trustee", strconv.FormatBool(r.TrusteeService)},
		{"Period", strconv.Itoa(r.Period)},
	}
}

func (d *domainRegistrationOutput) JSONData() any { return d }

func (d *domainRegistrationOutput) TemplateData() any { return d.Data }

// domainRenewalOutput adapts DomainRenewal for output.
type domainRenewalOutput struct {
	Data *dnsimple.DomainRenewal `json:"data"`
}

func (d *domainRenewalOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (d *domainRenewalOutput) TableRows() [][]string {
	r := d.Data
	return [][]string{
		{"ID", strconv.FormatInt(r.ID, 10)},
		{"Domain ID", strconv.FormatInt(r.DomainID, 10)},
		{"State", r.State},
		{"Period", strconv.Itoa(r.Period)},
	}
}

func (d *domainRenewalOutput) JSONData() any { return d }

func (d *domainRenewalOutput) TemplateData() any { return d.Data }

// domainTransferOutput adapts DomainTransfer for output.
type domainTransferOutput struct {
	Data *dnsimple.DomainTransfer `json:"data"`
}

func (d *domainTransferOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (d *domainTransferOutput) TableRows() [][]string {
	t := d.Data
	return [][]string{
		{"ID", strconv.FormatInt(t.ID, 10)},
		{"Domain ID", strconv.FormatInt(t.DomainID, 10)},
		{"Registrant ID", strconv.FormatInt(t.RegistrantID, 10)},
		{"State", t.State},
		{"Auto Renew", strconv.FormatBool(t.AutoRenew)},
		{"WHOIS Privacy", strconv.FormatBool(t.WhoisPrivacy)},
		{"Trustee", strconv.FormatBool(t.TrusteeService)},
	}
}

func (d *domainTransferOutput) JSONData() any { return d }

func (d *domainTransferOutput) TemplateData() any { return d.Data }

func newRegistrarCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registrar",
		Short: "Domain registration operations",
	}

	cmd.AddCommand(newRegistrarCheckCmd(f))
	cmd.AddCommand(newRegistrarPricesCmd(f))
	cmd.AddCommand(newRegistrarRegisterCmd(f))
	cmd.AddCommand(newRegistrarTransferCmd(f))
	cmd.AddCommand(newRegistrarTransferOutCmd(f))
	cmd.AddCommand(newRegistrarRenewCmd(f))
	cmd.AddCommand(newRegistrarRestoreCmd(f))
	cmd.AddCommand(newRegistrarDelegationCmd(f))
	cmd.AddCommand(newRegistrarWhoisPrivacyCmd(f))
	cmd.AddCommand(newRegistrarAutoRenewalCmd(f))
	cmd.AddCommand(newRegistrarTransferLockCmd(f))
	cmd.AddCommand(newRegistrarRegistrantChangeCmd(f))

	return cmd
}

func newRegistrarCheckCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "check <domain>",
		Short: "Check domain availability",
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

			resp, err := c.Registrar.CheckDomain(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainCheckOutput{Data: resp.Data})
		},
	}
}

func newRegistrarPricesCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "prices <domain>",
		Short: "Get domain prices",
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

			resp, err := c.Registrar.GetDomainPrices(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainPriceOutput{Data: resp.Data})
		},
	}
}

func newRegistrarRegisterCmd(f *cmdutil.Factory) *cobra.Command {
	var registrantID int
	var autoRenew, whoisPrivacy, trusteeService bool
	var premiumPrice string
	var extendedAttributes []string

	cmd := &cobra.Command{
		Use:   "register <domain>",
		Short: "Register a domain",
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

			input := &dnsimple.RegisterDomainInput{
				RegistrantID:       registrantID,
				EnableAutoRenewal:  autoRenew,
				EnableWhoisPrivacy: whoisPrivacy,
				TrusteeService:     &trusteeService,
				PremiumPrice:       premiumPrice,
				ExtendedAttributes: parseExtendedAttributes(extendedAttributes),
			}

			resp, err := c.Registrar.RegisterDomain(context.Background(), accountID, args[0], input)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainRegistrationOutput{Data: resp.Data})
		},
	}

	cmd.Flags().IntVar(&registrantID, "registrant-id", 0, "Contact ID to use as registrant")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", true, "Enable auto-renewal")
	cmd.Flags().BoolVar(&whoisPrivacy, "whois-privacy", false, "Enable WHOIS privacy")
	cmd.Flags().BoolVar(&trusteeService, "trustee", false, "Enable trustee service (extra cost may apply)")
	cmd.Flags().StringVar(&premiumPrice, "premium-price", "", "Confirm premium price")
	cmd.Flags().StringArrayVar(&extendedAttributes, "extended-attribute", nil, "Extended attributes (key=value)")
	_ = cmd.MarkFlagRequired("registrant-id")

	return cmd
}

func newRegistrarTransferCmd(f *cmdutil.Factory) *cobra.Command {
	var registrantID int
	var authCode string
	var autoRenew, whoisPrivacy, trusteeService bool
	var premiumPrice string
	var extendedAttributes []string

	cmd := &cobra.Command{
		Use:   "transfer <domain>",
		Short: "Transfer a domain in",
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

			input := &dnsimple.TransferDomainInput{
				RegistrantID:       registrantID,
				AuthCode:           authCode,
				EnableAutoRenewal:  autoRenew,
				EnableWhoisPrivacy: whoisPrivacy,
				TrusteeService:     &trusteeService,
				PremiumPrice:       premiumPrice,
				ExtendedAttributes: parseExtendedAttributes(extendedAttributes),
			}

			resp, err := c.Registrar.TransferDomain(context.Background(), accountID, args[0], input)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainTransferOutput{Data: resp.Data})
		},
	}

	cmd.Flags().IntVar(&registrantID, "registrant-id", 0, "Contact ID to use as registrant")
	cmd.Flags().StringVar(&authCode, "auth-code", "", "Authorization code from current registrar")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", true, "Enable auto-renewal")
	cmd.Flags().BoolVar(&whoisPrivacy, "whois-privacy", false, "Enable WHOIS privacy")
	cmd.Flags().BoolVar(&trusteeService, "trustee", false, "Enable trustee service (extra cost may apply)")
	cmd.Flags().StringVar(&premiumPrice, "premium-price", "", "Confirm premium price")
	cmd.Flags().StringArrayVar(&extendedAttributes, "extended-attribute", nil, "Extended attributes (key=value)")
	_ = cmd.MarkFlagRequired("registrant-id")

	return cmd
}

func newRegistrarRenewCmd(f *cmdutil.Factory) *cobra.Command {
	var period int
	var premiumPrice string

	cmd := &cobra.Command{
		Use:   "renew <domain>",
		Short: "Renew a domain",
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

			input := &dnsimple.RenewDomainInput{
				Period:       period,
				PremiumPrice: premiumPrice,
			}

			resp, err := c.Registrar.RenewDomain(context.Background(), accountID, args[0], input)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainRenewalOutput{Data: resp.Data})
		},
	}

	cmd.Flags().IntVar(&period, "period", 1, "Number of years to renew")
	cmd.Flags().StringVar(&premiumPrice, "premium-price", "", "Confirm premium price")

	return cmd
}

func newRegistrarRestoreCmd(f *cmdutil.Factory) *cobra.Command {
	var premiumPrice string

	cmd := &cobra.Command{
		Use:   "restore <domain>",
		Short: "Restore an expired domain",
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

			input := &dnsimple.RenewDomainInput{
				PremiumPrice: premiumPrice,
			}

			resp, err := c.Registrar.RestoreDomain(context.Background(), accountID, args[0], input)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&domainRenewalOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&premiumPrice, "premium-price", "", "Confirm premium price")

	return cmd
}

func newRegistrarTransferOutCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer-out <domain>",
		Short: "Authorize a domain transfer out",
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

			_, err = c.Registrar.TransferDomainOut(context.Background(), accountID, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Transfer out authorized for %s\n", args[0])
			return nil
		},
	}
}

func parseExtendedAttributes(attrs []string) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, kv := range attrs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// certList adapts []Certificate for output.
type certList struct {
	Data       []dnsimple.Certificate `json:"data"`
	Pagination *dnsimple.Pagination   `json:"pagination,omitempty"`
}

func (c *certList) TableHeaders() []string {
	return []string{"ID", "COMMON NAME", "STATE", "AUTO RENEW", "EXPIRES AT"}
}

func (c *certList) TableRows() [][]string {
	rows := make([][]string, len(c.Data))
	for i, cert := range c.Data {
		rows[i] = []string{
			strconv.FormatInt(cert.ID, 10),
			cert.CommonName,
			cert.State,
			strconv.FormatBool(cert.AutoRenew),
			cert.ExpiresAt,
		}
	}
	return rows
}

func (c *certList) JSONData() any { return c }

func (c *certList) TemplateData() any { return c.Data }

// certItem adapts a single Certificate for output.
type certItem struct {
	Data *dnsimple.Certificate `json:"data"`
}

func (c *certItem) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (c *certItem) TableRows() [][]string {
	cert := c.Data
	rows := [][]string{
		{"ID", strconv.FormatInt(cert.ID, 10)},
		{"Common Name", cert.CommonName},
		{"State", cert.State},
		{"Authority", cert.AuthorityIdentifier},
		{"Auto Renew", strconv.FormatBool(cert.AutoRenew)},
		{"Years", strconv.Itoa(cert.Years)},
		{"Expires At", cert.ExpiresAt},
		{"Created At", cert.CreatedAt},
	}
	if len(cert.AlternateNames) > 0 {
		rows = append(rows, []string{"Alternate Names", strings.Join(cert.AlternateNames, ", ")})
	}
	return rows
}

func (c *certItem) JSONData() any { return c }

func (c *certItem) TemplateData() any { return c.Data }

// certBundleOutput adapts CertificateBundle for output.
type certBundleOutput struct {
	Data *dnsimple.CertificateBundle `json:"data"`
}

func (c *certBundleOutput) TableHeaders() []string {
	return []string{"COMPONENT", "VALUE"}
}

func (c *certBundleOutput) TableRows() [][]string {
	var rows [][]string
	if c.Data.ServerCertificate != "" {
		rows = append(rows, []string{"Server Certificate", c.Data.ServerCertificate})
	}
	if c.Data.RootCertificate != "" {
		rows = append(rows, []string{"Root Certificate", c.Data.RootCertificate})
	}
	for i, chain := range c.Data.IntermediateCertificates {
		rows = append(rows, []string{fmt.Sprintf("Intermediate %d", i+1), chain})
	}
	if c.Data.PrivateKey != "" {
		rows = append(rows, []string{"Private Key", c.Data.PrivateKey})
	}
	return rows
}

func (c *certBundleOutput) JSONData() any { return c }

func (c *certBundleOutput) TemplateData() any { return c.Data }

// certPurchaseOutput adapts CertificatePurchase for output.
type certPurchaseOutput struct {
	Data *dnsimple.CertificatePurchase `json:"data"`
}

func (c *certPurchaseOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (c *certPurchaseOutput) TableRows() [][]string {
	p := c.Data
	return [][]string{
		{"ID", strconv.FormatInt(p.ID, 10)},
		{"Certificate ID", strconv.FormatInt(p.CertificateID, 10)},
		{"State", p.State},
		{"Auto Renew", strconv.FormatBool(p.AutoRenew)},
	}
}

func (c *certPurchaseOutput) JSONData() any { return c }

func (c *certPurchaseOutput) TemplateData() any { return c.Data }

// certRenewalOutput adapts CertificateRenewal for output.
type certRenewalOutput struct {
	Data *dnsimple.CertificateRenewal `json:"data"`
}

func (c *certRenewalOutput) TableHeaders() []string {
	return []string{"FIELD", "VALUE"}
}

func (c *certRenewalOutput) TableRows() [][]string {
	r := c.Data
	return [][]string{
		{"ID", strconv.FormatInt(r.ID, 10)},
		{"Old Certificate ID", strconv.FormatInt(r.OldCertificateID, 10)},
		{"New Certificate ID", strconv.FormatInt(r.NewCertificateID, 10)},
		{"State", r.State},
		{"Auto Renew", strconv.FormatBool(r.AutoRenew)},
	}
}

func (c *certRenewalOutput) JSONData() any { return c }

func (c *certRenewalOutput) TemplateData() any { return c.Data }

func newCertificatesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificates",
		Short:   "Manage SSL/TLS certificates",
		Aliases: []string{"certs"},
	}

	cmd.AddCommand(newCertsListCmd(f))
	cmd.AddCommand(newCertsGetCmd(f))
	cmd.AddCommand(newCertsDownloadCmd(f))
	cmd.AddCommand(newCertsPrivateKeyCmd(f))
	cmd.AddCommand(newCertsLetsencryptCmd(f))

	return cmd
}

func newCertsListCmd(f *cmdutil.Factory) *cobra.Command {
	var sort string
	lf := &listFlags{}

	cmd := &cobra.Command{
		Use:   "list <domain>",
		Short: "List certificates for a domain",
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

			opts := &dnsimple.ListOptions{}
			if sort != "" {
				opts.Sort = &sort
			}

			return runList(cmd, f, lf, "certificates",
				func(page, perPage int) ([]dnsimple.Certificate, *dnsimple.Pagination, error) {
					if page > 0 {
						opts.Page = &page
					}
					if perPage > 0 {
						opts.PerPage = &perPage
					}
					resp, err := c.Certificates.ListCertificates(context.Background(), accountID, args[0], opts)
					if err != nil {
						return nil, nil, err
					}
					return resp.Data, resp.Pagination, nil
				},
				func(items []dnsimple.Certificate, pg *dnsimple.Pagination) *certList {
					return &certList{Data: items, Pagination: pg}
				})
		},
	}

	cmd.Flags().StringVar(&sort, "sort", "", "Sort order")
	lf.register(cmd)

	return cmd
}

func newCertsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <domain> <certificate-id>",
		Short: "Get certificate details",
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

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			resp, err := c.Certificates.GetCertificate(context.Background(), accountID, args[0], certID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certItem{Data: resp.Data})
		},
	}
}

func newCertsDownloadCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "download <domain> <certificate-id>",
		Short: "Download certificate bundle",
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

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			resp, err := c.Certificates.DownloadCertificate(context.Background(), accountID, args[0], certID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certBundleOutput{Data: resp.Data})
		},
	}
}

func newCertsPrivateKeyCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "private-key <domain> <certificate-id>",
		Short: "Get certificate private key",
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

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			resp, err := c.Certificates.GetCertificatePrivateKey(context.Background(), accountID, args[0], certID)
			if err != nil {
				return err
			}

			// For private key, just output the PEM directly in table mode
			if !f.Flags.JSON && f.Flags.Format == "" {
				fmt.Fprint(cmd.OutOrStdout(), resp.Data.PrivateKey)
				return nil
			}
			return f.Printer(cmd).Print(&certBundleOutput{Data: resp.Data})
		},
	}
}

func newCertsLetsencryptCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "letsencrypt",
		Short: "Manage Let's Encrypt certificates",
	}

	cmd.AddCommand(newLetsencryptPurchaseCmd(f))
	cmd.AddCommand(newLetsencryptIssueCmd(f))
	cmd.AddCommand(newLetsencryptRenewCmd(f))
	cmd.AddCommand(newLetsencryptIssueRenewalCmd(f))

	return cmd
}

func newLetsencryptPurchaseCmd(f *cmdutil.Factory) *cobra.Command {
	var name string
	var autoRenew bool
	var alternateNames []string
	var signatureAlgorithm string

	cmd := &cobra.Command{
		Use:   "purchase <domain>",
		Short: "Purchase a Let's Encrypt certificate",
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

			attrs := dnsimple.LetsencryptCertificateAttributes{
				Name:               name,
				AutoRenew:          autoRenew,
				AlternateNames:     alternateNames,
				SignatureAlgorithm: signatureAlgorithm,
			}

			resp, err := c.Certificates.PurchaseLetsencryptCertificate(context.Background(), accountID, args[0], attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certPurchaseOutput{Data: resp.Data})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Certificate name (subdomain)")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Enable auto-renewal")
	cmd.Flags().StringSliceVar(&alternateNames, "alternate-names", nil, "Subject Alternative Names")
	cmd.Flags().StringVar(&signatureAlgorithm, "signature-algorithm", "", "Signature algorithm")

	return cmd
}

func newLetsencryptIssueCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "issue <domain> <certificate-id>",
		Short: "Issue a pending Let's Encrypt certificate",
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

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			resp, err := c.Certificates.IssueLetsencryptCertificate(context.Background(), accountID, args[0], certID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certItem{Data: resp.Data})
		},
	}
}

func newLetsencryptRenewCmd(f *cmdutil.Factory) *cobra.Command {
	var autoRenew bool
	var signatureAlgorithm string

	cmd := &cobra.Command{
		Use:   "renew <domain> <certificate-id>",
		Short: "Purchase a Let's Encrypt certificate renewal",
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

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			attrs := dnsimple.LetsencryptCertificateAttributes{
				AutoRenew:          autoRenew,
				SignatureAlgorithm: signatureAlgorithm,
			}

			resp, err := c.Certificates.PurchaseLetsencryptCertificateRenewal(context.Background(), accountID, args[0], certID, attrs)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certRenewalOutput{Data: resp.Data})
		},
	}

	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Enable auto-renewal")
	cmd.Flags().StringVar(&signatureAlgorithm, "signature-algorithm", "", "Signature algorithm")

	return cmd
}

func newLetsencryptIssueRenewalCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "issue-renewal <domain> <certificate-id> <renewal-id>",
		Short: "Issue a pending Let's Encrypt certificate renewal",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			certID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid certificate ID: %s", args[1])
			}

			renewalID, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid renewal ID: %s", args[2])
			}

			resp, err := c.Certificates.IssueLetsencryptCertificateRenewal(context.Background(), accountID, args[0], certID, renewalID)
			if err != nil {
				return err
			}

			return f.Printer(cmd).Print(&certItem{Data: resp.Data})
		},
	}
}

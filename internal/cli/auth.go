package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-cli/internal/client"
	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

func newAuthCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with DNSimple",
	}

	cmd.AddCommand(newAuthLoginCmd(f))
	cmd.AddCommand(newAuthLogoutCmd(f))
	cmd.AddCommand(newAuthStatusCmd(f))
	cmd.AddCommand(newAuthSwitchCmd(f))

	return cmd
}

func newAuthLoginCmd(f *cmdutil.Factory) *cobra.Command {
	var withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a DNSimple API token",
		Long: `Authenticate with DNSimple by providing an API token.

Get your token from:

  Production: https://dnsimple.com/user
  Sandbox:    https://sandbox.dnsimple.com/user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var token string

			if withToken {
				// Read token from stdin (for piping)
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if token == "" {
					return fmt.Errorf("no token provided on stdin")
				}
			} else {
				fmt.Fprint(os.Stderr, "Paste your API token: ")
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if token == "" {
					return fmt.Errorf("no token provided")
				}
			}

			// Validate the token by calling Whoami
			c := client.NewClient(cfg, token)
			whoami, err := c.Identity.Whoami(context.Background())
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			cred := &config.HostCredential{
				Token: token,
			}

			// Determine account
			if whoami.Data.Account != nil {
				// Account token — single account
				cred.AccountID = strconv.FormatInt(whoami.Data.Account.ID, 10)
				if whoami.Data.User != nil {
					cred.User = whoami.Data.User.Email
				}
			} else if whoami.Data.User != nil {
				// User token — may have multiple accounts
				cred.User = whoami.Data.User.Email

				accounts, err := c.Accounts.ListAccounts(context.Background(), nil)
				if err != nil {
					return fmt.Errorf("failed to list accounts: %w", err)
				}

				if len(accounts.Data) == 1 {
					cred.AccountID = strconv.FormatInt(accounts.Data[0].ID, 10)
				} else if len(accounts.Data) > 1 {
					cred.AccountID, err = promptForAccountSelection(os.Stdin, os.Stderr, accounts.Data)
					if err != nil {
						return err
					}
				}
			}

			// Store credentials
			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			host := cfg.HostKey()
			creds.Set(host, cred)
			if err := creds.Save(); err != nil {
				return err
			}

			env := "production"
			if cfg.Sandbox {
				env = "sandbox"
			}

			fmt.Fprintf(os.Stderr, "Logged in to %s as %s (account: %s)\n", env, cred.User, cred.AccountID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")

	return cmd
}

func newAuthLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored authentication credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			host := cfg.HostKey()
			creds.Delete(host)
			if err := creds.Save(); err != nil {
				return err
			}

			env := "production"
			if cfg.Sandbox {
				env = "sandbox"
			}
			fmt.Fprintf(os.Stderr, "Logged out of %s\n", env)
			return nil
		},
	}
}

func promptForAccountSelection(in io.Reader, errOut io.Writer, accounts []dnsimple.Account) (string, error) {
	fmt.Fprintln(errOut, "")
	fmt.Fprintln(errOut, "Multiple accounts available:")
	fmt.Fprintln(errOut, "")
	for i, a := range accounts {
		fmt.Fprintf(errOut, "  [%d] %s (ID: %d)\n", i+1, a.Email, a.ID)
	}
	fmt.Fprintln(errOut, "")
	fmt.Fprint(errOut, "Select account number: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read account selection: %w", err)
		}
		return "", fmt.Errorf("no account selected")
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(accounts) {
		return "", fmt.Errorf("invalid selection")
	}

	return strconv.FormatInt(accounts[choice-1].ID, 10), nil
}

func newAuthStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			whoami, err := c.Identity.Whoami(context.Background())
			if err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			env := "production"
			if cfg.Sandbox {
				env = "sandbox"
			}

			fmt.Fprintf(os.Stdout, "Environment: %s\n", env)
			if whoami.Data.User != nil {
				fmt.Fprintf(os.Stdout, "User:        %s (ID: %d)\n", whoami.Data.User.Email, whoami.Data.User.ID)
			}
			if whoami.Data.Account != nil {
				fmt.Fprintf(os.Stdout, "Account:     %s (ID: %d)\n", whoami.Data.Account.Email, whoami.Data.Account.ID)
			}

			return nil
		},
	}
}

func newAuthSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <account-id>",
		Short: "Switch default account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			host := cfg.HostKey()
			cred := creds.Get(host)
			if cred == nil {
				return fmt.Errorf("not authenticated. Run 'dnsimple auth login' first")
			}

			cred.AccountID = args[0]
			creds.Set(host, cred)
			if err := creds.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Switched to account %s\n", args[0])
			return nil
		},
	}
}

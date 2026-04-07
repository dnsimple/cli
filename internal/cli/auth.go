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

type authStatusOutput struct {
	Environment  string `json:"environment"`
	UserID       int64  `json:"user_id,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	AccountID    int64  `json:"account_id,omitempty"`
	AccountEmail string `json:"account_email,omitempty"`
	Warning      string `json:"warning,omitempty"`
}

func (a *authStatusOutput) TableHeaders() []string {
	return []string{"KEY", "VALUE"}
}

func (a *authStatusOutput) TableRows() [][]string {
	rows := [][]string{{"Environment", a.Environment}}
	if a.UserID != 0 {
		rows = append(rows, []string{"User", fmt.Sprintf("%s (ID: %d)", a.UserEmail, a.UserID)})
	}
	if a.AccountID != 0 {
		rows = append(rows, []string{"Account", fmt.Sprintf("%s (ID: %d)", a.AccountEmail, a.AccountID)})
	}
	if a.Warning != "" {
		rows = append(rows, []string{"Warning", a.Warning})
	}
	return rows
}

func (a *authStatusOutput) JSONData() any {
	return a
}

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
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a DNSimple API token",
		Long: `Authenticate with DNSimple by providing an API token and store it as a named context.

The new context becomes the active one. To create a sandbox context, pass --sandbox.
To choose a context name, pass --name; otherwise the name is derived from the
environment ('production' or 'sandbox'), with the account ID appended on collision.

Get your token from:

  Production: https://dnsimple.com/user
  Sandbox:    https://sandbox.dnsimple.com/user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host := config.HostForSandbox(f.Flags.Sandbox)

			token, err := readLoginToken(cmd, withToken)
			if err != nil {
				return err
			}

			// Validate the token by calling Whoami against the chosen host.
			// cfg.BaseURL is set by the factory from the --sandbox flag, and
			// can be stubbed by tests via f.Config.
			c := client.New(client.Options{
				BaseURL: cfg.BaseURL,
				Token:   token,
				Version: f.Version,
			})
			whoami, err := c.Identity.Whoami(context.Background())
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			accountID, user, err := resolveLoginAccount(c, whoami, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			ctx, action, err := upsertLoginContext(creds, host, token, accountID, user, nameFlag)
			if err != nil {
				return err
			}

			creds.ActiveContext = ctx.Name
			if err := creds.Save(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "%s context %q (%s, account %s) and set as active\n",
				action, ctx.Name, config.EnvironmentName(host), ctx.AccountID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Name for the new context (auto-derived if omitted)")

	return cmd
}

// readLoginToken reads a token either from the command's stdin (when
// --with-token is set or when reading interactively) and trims whitespace.
func readLoginToken(cmd *cobra.Command, withToken bool) (string, error) {
	if !withToken {
		fmt.Fprint(cmd.ErrOrStderr(), "Paste your API token: ")
	}
	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		if withToken {
			return "", fmt.Errorf("no token provided on stdin")
		}
		return "", fmt.Errorf("no token provided")
	}
	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		if withToken {
			return "", fmt.Errorf("no token provided on stdin")
		}
		return "", fmt.Errorf("no token provided")
	}
	return token, nil
}

// resolveLoginAccount determines the account ID and user email associated
// with the freshly-authenticated token, prompting the user when a multi-
// account user token is presented.
func resolveLoginAccount(c *dnsimple.Client, whoami *dnsimple.WhoamiResponse, in io.Reader, errOut io.Writer) (string, string, error) {
	if whoami.Data.Account != nil {
		accountID := strconv.FormatInt(whoami.Data.Account.ID, 10)
		user := ""
		if whoami.Data.User != nil {
			user = whoami.Data.User.Email
		}
		return accountID, user, nil
	}

	if whoami.Data.User == nil {
		return "", "", fmt.Errorf("token has no associated user or account")
	}
	user := whoami.Data.User.Email

	accounts, err := c.Accounts.ListAccounts(context.Background(), nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to list accounts: %w", err)
	}
	switch len(accounts.Data) {
	case 0:
		return "", user, nil
	case 1:
		return strconv.FormatInt(accounts.Data[0].ID, 10), user, nil
	default:
		accountID, err := promptForAccountSelection(in, errOut, accounts.Data)
		if err != nil {
			return "", "", err
		}
		return accountID, user, nil
	}
}

// upsertLoginContext finds or creates a context for the given login.
//
// With an explicit name:
//   - free name → create.
//   - same (host, token) under that name → refresh metadata.
//   - same (host, token) under a different name → reject (token already stored).
//   - different (host, token) under that name → reject (name conflict).
//
// Without an explicit name:
//   - same (host, token) anywhere → refresh that context (re-login).
//   - otherwise → create with an auto-derived name.
//
// The returned action is "Created" or "Refreshed" for use in the success
// message.
func upsertLoginContext(creds *config.Credentials, host, token, accountID, user, explicitName string) (*config.Context, string, error) {
	if explicitName != "" {
		existing := creds.Find(explicitName)

		if existing != nil && existing.Host == host && existing.Token == token {
			existing.AccountID = accountID
			existing.User = user
			return existing, "Refreshed", nil
		}
		if existing != nil {
			return nil, "", fmt.Errorf("context %q already exists; pick a different --name or run 'dnsimple auth logout --name %s' first", explicitName, explicitName)
		}
		// Token already stored under another name?
		for _, ctx := range creds.Contexts {
			if ctx.Host == host && ctx.Token == token {
				return nil, "", fmt.Errorf("this token is already stored as context %q; run 'dnsimple auth logout --name %s' first if you want to use a different name", ctx.Name, ctx.Name)
			}
		}
		newCtx := &config.Context{
			Name:      explicitName,
			Host:      host,
			Token:     token,
			AccountID: accountID,
			User:      user,
		}
		creds.Add(newCtx)
		return newCtx, "Created", nil
	}

	// Auto-derived path. Re-login detection first.
	for _, ctx := range creds.Contexts {
		if ctx.Host == host && ctx.Token == token {
			ctx.AccountID = accountID
			ctx.User = user
			return ctx, "Refreshed", nil
		}
	}

	name := pickAutoContextName(creds, host, accountID)
	newCtx := &config.Context{
		Name:      name,
		Host:      host,
		Token:     token,
		AccountID: accountID,
		User:      user,
	}
	creds.Add(newCtx)
	return newCtx, "Created", nil
}

// pickAutoContextName picks the auto-derived context name following the
// algorithm in dnsimple/dnsimple-cli#28:
//
//  1. Bare environment name (production or sandbox)
//  2. <env>-<account_id>
//  3. <env>-<account_id>-N for N=2, 3, ...
//
// Re-login detection (step 0) is handled by the caller before this function
// is invoked.
func pickAutoContextName(creds *config.Credentials, host, accountID string) string {
	env := config.EnvironmentName(host)

	if creds.Find(env) == nil {
		return env
	}

	base := env
	if accountID != "" {
		base = env + "-" + accountID
	}
	if creds.Find(base) == nil {
		return base
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if creds.Find(candidate) == nil {
			return candidate
		}
	}
}


func newAuthLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a stored authentication context",
		Long: `Remove a stored authentication context.

Without --name, removes the active context. If the active context is removed,
the active selection shifts to the first remaining context, or is cleared if
no contexts remain.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			target := nameFlag
			if target == "" {
				target = creds.ActiveContext
			}
			if target == "" {
				return fmt.Errorf("no context to remove (no active context, no --name)")
			}

			if !creds.Remove(target) {
				return fmt.Errorf("context %q not found", target)
			}
			if err := creds.Save(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Removed context %q\n", target)
			switch {
			case len(creds.Contexts) == 0:
				fmt.Fprintln(cmd.ErrOrStderr(), "No contexts remain.")
			case creds.ActiveContext != "":
				fmt.Fprintf(cmd.ErrOrStderr(), "Active context is now %q\n", creds.ActiveContext)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Name of the context to remove (defaults to active)")

	return cmd
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

			data := &authStatusOutput{Environment: env}
			if whoami.Data.User != nil {
				data.UserID = whoami.Data.User.ID
				data.UserEmail = whoami.Data.User.Email
			}

			// Show the default account commands will actually use, resolved through
			// the same precedence chain (flag → env → config → stored credentials).
			// Enrich the ID with an email by looking it up in the accessible accounts,
			// and warn if the stored default no longer matches anything the token can see.
			if accountID, err := f.AccountID(); err == nil && accountID != "" {
				if id, parseErr := strconv.ParseInt(accountID, 10, 64); parseErr == nil {
					data.AccountID = id
					if accounts, err := c.Accounts.ListAccounts(context.Background(), nil); err == nil {
						found := false
						for _, a := range accounts.Data {
							if a.ID == id {
								data.AccountEmail = a.Email
								found = true
								break
							}
						}
						if !found {
							data.Warning = fmt.Sprintf("account %d is not accessible with the current token; run 'dnsimple auth switch <account-id>' to update", id)
						}
					}
				}
			}

			return f.Printer(cmd).Print(data)
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

			// Validate the requested account is accessible with the current token.
			c, err := f.Client()
			if err != nil {
				return err
			}
			accounts, err := c.Accounts.ListAccounts(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("failed to list accounts: %w", err)
			}

			target := args[0]
			found := false
			for _, a := range accounts.Data {
				if strconv.FormatInt(a.ID, 10) == target {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("account %s is not accessible with the current token", target)
			}

			cred.AccountID = target
			creds.Set(host, cred)
			if err := creds.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Switched to account %s\n", target)
			return nil
		},
	}
}

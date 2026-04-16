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
	"golang.org/x/term"
)

type authStatusOutput struct {
	Context      string `json:"context,omitempty"`
	Environment  string `json:"environment"`
	Host         string `json:"host,omitempty"`
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
	var rows [][]string
	if a.Context != "" {
		rows = append(rows, []string{"Context", a.Context})
	}
	rows = append(rows, []string{"Environment", a.Environment})
	if a.Host != "" {
		rows = append(rows, []string{"Host", a.Host})
	}
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

func (a *authStatusOutput) TemplateData() any {
	return a
}

func newAuthCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with DNSimple",
	}

	cmd.AddCommand(newAuthLoginCmd(f))
	cmd.AddCommand(newAuthLogoutCmd(f))
	cmd.AddCommand(newAuthListCmd(f))
	cmd.AddCommand(newAuthStatusCmd(f))
	cmd.AddCommand(newAuthSwitchCmd(f))

	return cmd
}

// authContextRow is a single row in the `auth list` output.
type authContextRow struct {
	Active      bool   `json:"active"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Host        string `json:"host"`
	AccountID   string `json:"account_id,omitempty"`
	User        string `json:"user,omitempty"`
}

// authContextList is the `auth list` output.
type authContextList struct {
	Data []authContextRow `json:"contexts"`
}

func (a *authContextList) TableHeaders() []string {
	return []string{"", "NAME", "ENVIRONMENT", "ACCOUNT", "USER"}
}

func (a *authContextList) TableRows() [][]string {
	rows := make([][]string, len(a.Data))
	for i, c := range a.Data {
		marker := ""
		if c.Active {
			marker = "*"
		}
		rows[i] = []string{marker, c.Name, c.Environment, c.AccountID, c.User}
	}
	return rows
}

func (a *authContextList) JSONData() any {
	return a
}

func (a *authContextList) TemplateData() any {
	return a.Data
}

func newAuthListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List authentication contexts",
		Long: `List all stored authentication contexts and highlight the active one.

This command does not contact the DNSimple API and works without a valid token.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}

			rows := make([]authContextRow, len(creds.Contexts))
			for i, ctx := range creds.Contexts {
				rows[i] = authContextRow{
					Active:      ctx.Name == creds.ActiveContext,
					Name:        ctx.Name,
					Environment: config.EnvironmentName(ctx.Host),
					Host:        ctx.Host,
					AccountID:   ctx.AccountID,
					User:        ctx.User,
				}
			}

			if len(rows) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No contexts. Run 'dnsimple auth login' to create one.")
			}

			return f.Printer(cmd).Print(&authContextList{Data: rows})
		},
	}
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
			host := config.HostForSandbox(cfg.Sandbox)

			token, err := readLoginToken(cmd, withToken)
			if err != nil {
				return err
			}

			// Validate the token by calling Whoami against the effective API
			// base URL. cfg.Sandbox already reflects persisted config/env plus
			// any --sandbox flag override, and cfg.BaseURL can be stubbed by
			// tests via f.Config.
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

// readLoginToken reads a token from the command's stdin and trims whitespace.
//
// With --with-token, input is read as a single line (the typical piping case)
// and is not masked.
//
// Interactive (no --with-token), when stdin is a real TTY, the input is read
// with terminal echo disabled so the token is not displayed. When stdin is
// not a real TTY (tests, redirected input), the function falls back to a
// plain line scan so behaviour stays predictable.
func readLoginToken(cmd *cobra.Command, withToken bool) (string, error) {
	if withToken {
		token, err := scanLine(cmd.InOrStdin())
		if err != nil || token == "" {
			return "", fmt.Errorf("no token provided on stdin")
		}
		return token, nil
	}

	fmt.Fprint(cmd.ErrOrStderr(), "Paste your API token: ")

	if f, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		raw, err := term.ReadPassword(int(f.Fd()))
		// ReadPassword leaves the cursor on the prompt line; emit a newline
		// so subsequent output starts cleanly.
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("failed to read token: %w", err)
		}
		token := strings.TrimSpace(string(raw))
		if token == "" {
			return "", fmt.Errorf("no token provided")
		}
		return token, nil
	}

	token, err := scanLine(cmd.InOrStdin())
	if err != nil || token == "" {
		return "", fmt.Errorf("no token provided")
	}
	return token, nil
}

// scanLine reads a single trimmed line from r.
func scanLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return strings.TrimSpace(scanner.Text()), nil
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

// promptForAccountSelection presents a numbered list of accounts during
// `auth login` when a user token grants access to more than one account.
//
// TUI: good candidate for an interactive picker (arrow keys, filtering) once
// the CLI takes on a TUI library as a global dependency. Until then, the
// numbered prompt is consistent with promptForContextSelection and works in
// non-TTY environments such as piped stdin and CI.
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
			rc, err := f.Context()
			if err != nil {
				return err
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			whoami, err := c.Identity.Whoami(context.Background())
			if err != nil {
				return err
			}

			data := &authStatusOutput{
				Context:     rc.ContextName,
				Environment: config.EnvironmentName(rc.Host),
				Host:        rc.Host,
			}
			if whoami.Data.User != nil {
				data.UserID = whoami.Data.User.ID
				data.UserEmail = whoami.Data.User.Email
			}

			// Enrich the resolved account ID with an email by looking it up in
			// the accessible accounts. If the lookup succeeds but the account is
			// not in the list, surface a warning so the user notices stale state.
			if rc.AccountID != "" {
				if id, parseErr := strconv.ParseInt(rc.AccountID, 10, 64); parseErr == nil {
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
							data.Warning = fmt.Sprintf("account %d is not accessible with the current token; run 'dnsimple auth login' to refresh the context", id)
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
		Use:   "switch [<name-or-account-id>]",
		Short: "Switch the active authentication context",
		Long: `Switch the active authentication context.

With an argument: switches the active context to the named context. If no
context with that name exists, falls back to matching against stored account
IDs (errors when ambiguous).

Without an argument: opens an interactive picker listing every stored context.

Switching is a local operation; it does not contact the DNSimple API.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}
			if len(creds.Contexts) == 0 {
				return fmt.Errorf("no contexts. Run 'dnsimple auth login' to create one")
			}

			var target string
			if len(args) == 1 {
				target, err = resolveSwitchTarget(creds, args[0])
				if err != nil {
					return err
				}
			} else {
				target, err = promptForContextSelection(cmd.InOrStdin(), cmd.ErrOrStderr(), creds)
				if err != nil {
					return err
				}
			}

			if target == creds.ActiveContext {
				fmt.Fprintf(cmd.ErrOrStderr(), "Already on context %q\n", target)
				return nil
			}

			if err := creds.SetActive(target); err != nil {
				return err
			}
			if err := creds.Save(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Switched to context %q\n", target)
			return nil
		},
	}
}

// resolveSwitchTarget maps the user's argument to a stored context name. The
// argument is matched first as a context name, then as an account ID. An
// account ID that matches multiple contexts produces an error listing them.
func resolveSwitchTarget(creds *config.Credentials, arg string) (string, error) {
	if creds.Find(arg) != nil {
		return arg, nil
	}

	var matches []*config.Context
	for _, ctx := range creds.Contexts {
		if ctx.AccountID == arg {
			matches = append(matches, ctx)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no context named %q (and no stored context has account %s)", arg, arg)
	case 1:
		return matches[0].Name, nil
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("multiple contexts have account %s (%s); pass the context name instead", arg, strings.Join(names, ", "))
	}
}

// promptForContextSelection renders a numbered list of contexts and reads the
// user's choice from stdin. The active context is marked with a "*".
//
// TUI: good candidate for an interactive picker (arrow keys, filtering) once
// the CLI takes on a TUI library as a global dependency. Until then, the
// numbered prompt is consistent with promptForAccountSelection, works in
// non-TTY environments, and is trivial to test.
func promptForContextSelection(in io.Reader, errOut io.Writer, creds *config.Credentials) (string, error) {
	fmt.Fprintln(errOut, "")
	fmt.Fprintln(errOut, "Available contexts:")
	fmt.Fprintln(errOut, "")
	for i, ctx := range creds.Contexts {
		marker := " "
		if ctx.Name == creds.ActiveContext {
			marker = "*"
		}
		fmt.Fprintf(errOut, "  %s [%d] %-20s %-12s %s (%s)\n",
			marker, i+1, ctx.Name, config.EnvironmentName(ctx.Host), ctx.User, ctx.AccountID)
	}
	fmt.Fprintln(errOut, "")
	fmt.Fprint(errOut, "Select context number: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read context selection: %w", err)
		}
		return "", fmt.Errorf("no context selected")
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(creds.Contexts) {
		return "", fmt.Errorf("invalid selection")
	}
	return creds.Contexts[choice-1].Name, nil
}

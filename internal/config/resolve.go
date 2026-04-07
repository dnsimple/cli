package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ResolvedContext is the fully-resolved authentication context for a single
// invocation of the CLI. It is the product of the resolution chain
// (raw flags/env → --context → active context).
type ResolvedContext struct {
	// ContextName is the name of the stored context the resolution drew from,
	// or empty if the invocation was satisfied entirely by raw overrides.
	ContextName string

	// BaseURL is the API URL to talk to.
	BaseURL string

	// Host is the canonical host name (e.g. api.dnsimple.com).
	Host string

	// Token is the API token to send.
	Token string

	// AccountID is the resolved account ID.
	AccountID string

	// User is the email or handle of the token owner, when known.
	User string
}

// ResolveOptions captures the per-invocation overrides used during resolution.
type ResolveOptions struct {
	// Token is the value of --token (empty if not set).
	Token string

	// Account is the value of --account (empty if not set).
	Account string

	// ContextName is the value of --context (empty if not set).
	ContextName string

	// Sandbox is the value of --sandbox.
	Sandbox bool

	// DefaultAccount is the cfg.DefaultAccount value (from config file).
	DefaultAccount string

	// BaseURLOverride is a test-only override that short-circuits host-based
	// BaseURL derivation. Production code never sets this.
	BaseURLOverride string
}

// Resolve produces a ResolvedContext from the given credentials and options.
//
// Resolution order, applied per field:
//
//  1. Explicit raw override (--token / --account / --sandbox)
//  2. Environment variable (DNSIMPLE_TOKEN / DNSIMPLE_ACCOUNT)
//  3. cfg.DefaultAccount (account only)
//  4. Stored context selected by --context, --sandbox, or active_context
//
// Storage is only consulted if --context is set or if the raw overrides do
// not supply enough to satisfy the invocation.
func Resolve(creds *Credentials, opts ResolveOptions) (*ResolvedContext, error) {
	rc := &ResolvedContext{}

	// Gather raw token / account overrides up front.
	rawToken := opts.Token
	if rawToken == "" {
		rawToken = os.Getenv("DNSIMPLE_TOKEN")
	}
	rawAccount := opts.Account
	if rawAccount == "" {
		rawAccount = os.Getenv("DNSIMPLE_ACCOUNT")
	}
	if rawAccount == "" {
		rawAccount = opts.DefaultAccount
	}

	// Decide whether we need to consult stored credentials.
	//
	// - If --context is set, always pick it (the user explicitly asked).
	// - If --token AND --account are both supplied raw, we can run fully
	//   stateless without touching storage.
	// - Otherwise, fall back to storage to fill the gaps.
	needsContext := opts.ContextName != "" || rawToken == "" || rawAccount == ""

	var ctx *Context
	if needsContext {
		var err error
		ctx, err = pickContext(creds, opts)
		if err != nil {
			return nil, err
		}
	}

	if ctx != nil {
		rc.ContextName = ctx.Name
		rc.Token = ctx.Token
		rc.AccountID = ctx.AccountID
		rc.Host = ctx.Host
		rc.User = ctx.User
	}

	// Apply raw overrides on top.
	if rawToken != "" {
		rc.Token = rawToken
	}
	if rawAccount != "" {
		rc.AccountID = rawAccount
	}

	// --sandbox host override wins over the context-supplied host.
	if opts.Sandbox {
		rc.Host = SandboxHost
	}
	if rc.Host == "" {
		rc.Host = ProductionHost
	}

	if rc.Token == "" {
		return nil, errors.New("not authenticated. Run 'dnsimple auth login' to authenticate")
	}
	if rc.AccountID == "" {
		return nil, errors.New("no account specified. Use --account flag, DNSIMPLE_ACCOUNT env var, or run 'dnsimple auth login'")
	}

	if opts.BaseURLOverride != "" {
		rc.BaseURL = opts.BaseURLOverride
	} else {
		rc.BaseURL = BaseURLForHost(rc.Host)
	}

	return rc, nil
}

// pickContext returns the stored context to draw defaults from, applying the
// --context, --sandbox, and active-context selection rules.
func pickContext(creds *Credentials, opts ResolveOptions) (*Context, error) {
	if opts.ContextName != "" {
		ctx := creds.Find(opts.ContextName)
		if ctx == nil {
			return nil, fmt.Errorf("context %q not found", opts.ContextName)
		}
		return ctx, nil
	}

	if opts.Sandbox {
		// Constrain selection to sandbox contexts.
		if active := creds.Active(); active != nil && active.Host == SandboxHost {
			return active, nil
		}
		sandboxes := contextsByHost(creds, SandboxHost)
		switch len(sandboxes) {
		case 0:
			return nil, nil
		case 1:
			return sandboxes[0], nil
		default:
			names := make([]string, len(sandboxes))
			for i, s := range sandboxes {
				names[i] = s.Name
			}
			return nil, fmt.Errorf(
				"multiple sandbox contexts in storage (%s); pass --context <name> to choose one or use --token/--account for raw credentials",
				strings.Join(names, ", "),
			)
		}
	}

	return creds.Active(), nil
}

func contextsByHost(creds *Credentials, host string) []*Context {
	var out []*Context
	for _, c := range creds.Contexts {
		if c.Host == host {
			out = append(out, c)
		}
	}
	return out
}

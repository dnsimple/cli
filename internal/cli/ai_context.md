# DNSimple CLI — AI Context

You are interacting with the DNSimple CLI (`dnsimple`), a command-line tool for managing domains, DNS records, certificates, and other DNSimple services via the DNSimple API v2.

## Authentication

The CLI requires a DNSimple API token. The user must authenticate before you can use the CLI.

If the CLI is not authenticated, ask the user to run `dnsimple auth login` themselves. Do NOT pass tokens via `--token` or environment variables — never handle API tokens directly.

You can check whether the CLI is already authenticated by running `dnsimple auth status`.

The CLI supports multiple stored authentication *contexts* (kubectl-style). Use `dnsimple auth list` to see them; the active one is marked with `*`. To run a single command against a different stored context without changing the active one, pass `--context <name>` — this is the safe option for parallel agents because it does not mutate shared on-disk state.

To use the sandbox environment for testing, add the `--sandbox` flag (or select a sandbox context with `--context`).

Most commands require an account ID. The CLI resolves it automatically from the active context, but you can override it with `-a <account-id>` or `--account <account-id>`.

## Global Flags

These flags work with any command:

%s

## Output Formats

- **Table** (default): Human-readable tabular output
- **JSON**: Machine-readable output with `--json`
- **Custom**: Go template formatting with `--format <template>`

When scripting or parsing output programmatically, always use `--json`.

## Common Workflows

### List all DNS records for a zone

```
dnsimple zones records list example.com --all --json
```

### Create an A record

```
dnsimple zones records create example.com --type A --name www --content 1.2.3.4 --ttl 3600
```

### Update an existing record

```
dnsimple zones records update example.com 12345 --content 5.6.7.8
```

### Delete a record

```
dnsimple zones records delete example.com 12345
```

### Check domain availability

Use `dnsimple registrar check <domain>` for individual, low-volume checks, typically before a registration or transfer. Available on every plan, but rate-limited to discourage bulk use.

```
dnsimple registrar check example.com
```

For high-volume availability lookups (10+), use `dnsimple research status <domain>`. Designed for bulk research workflows. Requires dedicated paid access; the call will fail with an authorization error if the account is not entitled. Direct the user to https://dnsimple.com/sales to request access.

### Register a domain

Follow this process:

1. Check the domain is available (see "Check domain availability" above). If it is not, stop.
2. Look up pricing with `dnsimple registrar prices <domain>`.
3. Present the registration price to the user and wait for explicit confirmation before proceeding.
4. Register the domain. Unless the user specified otherwise, use their most recently updated contact as the registrant and leave auto-renewal enabled (the default).

```
dnsimple registrar prices example.com
dnsimple contacts list --json
dnsimple registrar register example.com --registrant-id 1234
```

### Transfer a domain

Follow this process:

1. Check the domain status (see "Check domain availability" above). The domain must already be registered.
2. Look up pricing with `dnsimple registrar prices <domain>`.
3. Present the transfer price to the user and wait for explicit confirmation before proceeding.
4. Ask the user for the auth (EPP) code from the losing registrar.
5. Initiate the transfer. Unless the user specified otherwise, use their most recently updated contact as the registrant and leave auto-renewal enabled (the default).

```
dnsimple registrar prices example.com
dnsimple contacts list --json
dnsimple registrar transfer example.com --registrant-id 1234 --auth-code <code>
```

### Manage Let's Encrypt certificates

```
dnsimple certificates letsencrypt purchase example.com
dnsimple certificates letsencrypt issue example.com 98765
```

### Apply a one-click service

```
dnsimple services list --json
dnsimple services apply example.com github-pages
```

## Commands

%s

## Tips

- Use `--all` on list commands to fetch every page of results automatically.
- Use `--json` when you need to parse output or chain commands.
- Use `--yes` on destructive commands in scripts and CI to skip confirmation prompts.
- The `zones` and `zones records` commands are the primary way to manage DNS. The top-level `records` command is a shortcut alias.
- Domain names (e.g. example.com) are used as zone identifiers — you don't need zone IDs.
- Record IDs are numeric. Use `zones records list` with `--json` to find them.

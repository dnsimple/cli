# Changelog

This project uses [Semantic Versioning 2.0.0](http://semver.org/), the format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Unreleased

- Release artifacts are published only to `dnsimple/cli`. The `dnsimple/homebrew-tap` release mirror is removed, and the install scripts and the Homebrew formula now download from `dnsimple/cli`.
- The new release notice now prints the release page URL for every installation method, colors the version numbers, and separates itself from the command output with a blank line above and below. `--no-color` and the `NO_COLOR` environment variable turn the color off. The `DNSIMPLE_NO_UPDATE_CHECK` environment variable, which turns the check off, is now documented in the README.
- Table output puts the header row in bold, and the pagination hints that go with a multi-page list are faint. `--no-color` and the `NO_COLOR` environment variable turn the color off, and redirected output stays plain.
- The update check no longer runs when the `BUILD_NUMBER` or the `RUN_ID` environment variable is set, and it now requires both the standard output stream and the standard error stream to be a terminal.
- Table output is cut to the width of the terminal. A long value, for example a DNSKEY record, no longer pushes the columns after it off the screen. Redirected output, `--json`, `--format`, and the single-resource `get` commands still return the full value.

## 0.10.0 - 2026-06-15

### Changed

- `auth login` now defaults to the interactive browser login (OAuth) on a terminal. To authenticate with an API token instead, pass `--with-token` and paste it when prompted. The `--web` flag that opted into the browser flow, the `oauth_login` config setting, and the `DNSIMPLE_OAUTH_LOGIN` environment variable that gated the dark launch have all been removed. (dnsimple/cli#61)

## 0.9.1 - 2026-06-10

### Added

- `auth login` can authenticate in the browser via an interactive OAuth flow (OAuth 2.0 with PKCE and a loopback redirect). The feature is dark-launched and off by default: opt in per command with `--web`, or persistently by setting `oauth_login: true` in the config file (or `DNSIMPLE_OAUTH_LOGIN=1`). Without it, `auth login` keeps prompting for a pasted API token. (dnsimple/cli#57)

## 0.9.0 - 2026-05-25

### Added

- List commands now show a `Showing X of Y` pagination hint and expose `--all`/`--page`/`--per-page` consistently, making it obvious when results span multiple pages and how to retrieve the rest. (dnsimple/cli#51)
- Trustee service support: pass `--trustee` to `registrar register` and `registrar transfer` to enable the trustee service on domains that require a local presence, and the trustee state now surfaces in `domains get`, `registrar prices`, `registrar check`, and `tlds` output. (dnsimple/cli#37)

### Changed

- Upgraded the DNSimple Go client (`dnsimple-go`) from v8 to v9. (dnsimple/cli#37)

## 0.8.0 - 2026-05-14

### Changed

- `dnsimple ai` no longer prints the full per-command flag tables by default; only required flags are listed. Pass `--full` for the complete reference, or run `<command> --help` to discover flags on demand. This shrinks the default output to reduce token usage.
- Updated `dnsimple ai` context to reduce duplicated command references and clarify command selection, improving agent context learning and token usage.
- Renamed the Go module path from `github.com/dnsimple/dnsimple-cli` to `github.com/dnsimple/cli` following the GitHub repository rename.

## 0.7.0 - 2026-04-24

### Changed

- `dnsimple services apply` and `dnsimple services unapply` now take arguments as `<domain> <service>`, matching other domain-scoped commands. This is a breaking change from the previous `<service> <domain>` order. (dnsimple/cli#19)
- Simplified the `vanity-name-servers` command to `vanity-servers`.
- Improved discoverability of the `dnsimple ai` command.
- Removed singular aliases (`domain`, `zone`, `record`, `template`, `contact`, `cert`) from plural resource commands to reduce help-output clutter and align with the conventions used by other popular CLIs. The `certs` alias for `certificates` is kept as a length-shortening shortcut. (dnsimple/cli#35)

## 0.6.0 - 2026-04-22

### Added

- `dnsimple research status <domain>` checks domain availability through the Domain Research endpoint, designed for high-volume bulk lookups and requiring dedicated paid access. (dnsimple/cli#4)
- `dnsimple ai` now includes a "Choosing the Right Command" section that documents when to use `registrar check` (rate-limited, available on every plan) versus `research status` (bulk lookups, paid add-on). (dnsimple/cli#4)

### Fixed

- `dnsimple completion` now shows help text when invoked without a shell argument instead of returning `Error: accepts 1 arg(s), received 0`. Restructured as a parent command with `bash`, `zsh`, `fish`, and `powershell` subcommands. (dnsimple/cli#38)

## 0.5.2 - 2026-04-15

### Added

- Add a native Windows `install.ps1` installer and publish it alongside `install.sh`.

## 0.5.1 - 2026-04-08

### Changed

- Release pipeline now mints a short-lived GitHub App installation token to push Homebrew formula updates to `dnsimple/homebrew-tap`, replacing the previous personal access token. Commits are server-side signed by GitHub and satisfy the tap's verified-signature branch protection rule. (dnsimple/cli#34)

## 0.5.0 - 2026-04-08

### Added

- Named authentication contexts: `auth login` now stores credentials under a named context (kubectl-style), `auth list` shows all stored contexts with the active one highlighted, `auth switch` selects the active context by name or account ID (with an interactive picker when called with no arguments), and `auth logout --name <name>` removes a specific context. Multiple production and sandbox contexts can coexist. A new `--context <name>` global flag provides stateless per-invocation overrides: run a command against a stored context without changing the active selection, intended for agents and parallel shells. (dnsimple/cli#28)
- `auth login` masks token input on a TTY when reading interactively. (dnsimple/cli#28)

### Changed

- Credentials file (`credentials.yml`) now stores a list of named contexts instead of a host-keyed map. Existing files using the legacy `hosts:` schema are migrated automatically and in place on first load. (dnsimple/cli#28)
- Field-by-field resolution chain: `--token`, `--account`, `--sandbox`, and `--context` each fall back independently to environment variables and then to the active stored context, so a single override does not require respecifying the others. (dnsimple/cli#28)
- Resource-level destructive commands such as `delete` and `unapply` now prompt for confirmation in interactive terminals and require `--yes` in non-interactive use. (dnsimple/cli#29)

### Fixed

- `domains delete` now preflights the domain state and requires stronger acknowledgment before deleting a registered domain, because the operation downgrades it to `hosted` and permanently loses registration metadata. (dnsimple/cli#16)
- `--format` now evaluates templates against the underlying resource data for wrapper-backed commands such as `domains get` and `domains list`, so templates like `{{.Name}}` and `{{range .}}{{.Name}}{{end}}` work as expected. (dnsimple/cli#17)
- `auth switch` now validates that the requested account is accessible with the current token, instead of silently storing any value. (dnsimple/cli#25)
- `auth status` now reports the resolved default account (the one commands actually use) instead of the token's bound account from `whoami`. (dnsimple/cli#25)
- `auth status` now reflects the active context's environment when the active context is sandbox without `--sandbox` being passed on the command line. (dnsimple/cli#28)
- `analytics query` now renders only the columns that match the requested `--groupings` (plus `VOLUME`), validates supported grouping keys (`zone_name`, `date`) before calling the API, and rejects unknown values with a clear error. (dnsimple/cli#21)

## 0.4.0 - 2026-04-03

### Added

- Add update notification that informs users when a newer CLI version is available. (dnsimple/cli#10)

## 0.3.0 - 2026-04-03

### Added

- Add Homebrew formula publishing via `dnsimple/homebrew-tap`.

## 0.2.0 - 2026-04-03

### Added

- Install script for curl-based installation with checksum verification. (dnsimple/cli#9)

## 0.1.0 - 2026-04-02

Initial release.

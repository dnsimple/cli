# Changelog

This project uses [Semantic Versioning 2.0.0](http://semver.org/), the format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## 0.11.0 - 2026-08-24

- Release artifacts are published only to `dnsimple/cli`, and the `dnsimple/homebrew-tap` release mirror is removed. The install scripts and the Homebrew formula now download from `dnsimple/cli`. (dnsimple/cli#69)
- The new release notice now prints the release page URL for every installation method, colors the version numbers, and separates itself from the command output with blank lines. `--no-color` and the `NO_COLOR` environment variable turn the color off. (dnsimple/cli#70)
- Table output puts the header row in bold and prints the pagination hints in faint color. `--no-color` and the `NO_COLOR` environment variable turn the color off, and redirected output stays plain. (dnsimple/cli#71)
- The update check does not run when the `BUILD_NUMBER` or the `RUN_ID` environment variable is set. It also requires both the standard output stream and the standard error stream to be a terminal. (dnsimple/cli#71)
- Table output is cut to the width of the terminal, so a long value, for example a DNSKEY record, does not push the columns after it off the screen. Redirected output, `--json`, `--format`, and the single-resource `get` commands still return the full value. (dnsimple/cli#72)

## 0.10.0 - 2026-06-15

### Changed

- `auth login` now defaults to the interactive browser login (OAuth) on a terminal; pass `--with-token` to authenticate with an API token instead. The `--web` flag, the `oauth_login` config setting, and the `DNSIMPLE_OAUTH_LOGIN` environment variable are removed. (dnsimple/cli#61)

## 0.9.1 - 2026-06-10

### Added

- `auth login` can authenticate in the browser via an interactive OAuth flow (OAuth 2.0 with PKCE and a loopback redirect). The feature is off by default: opt in with `--web`, with `oauth_login: true` in the config file, or with `DNSIMPLE_OAUTH_LOGIN=1`. (dnsimple/cli#57)

## 0.9.0 - 2026-05-25

### Added

- List commands now show a `Showing X of Y` pagination hint when results span multiple pages. All list commands expose `--all`, `--page`, and `--per-page`. (dnsimple/cli#51)
- Pass `--trustee` to `registrar register` and `registrar transfer` to enable the trustee service on domains that require a local presence. The trustee state shows in `domains get`, `registrar prices`, `registrar check`, and `tlds` output. (dnsimple/cli#37)

### Changed

- Upgrade the DNSimple Go client (`dnsimple-go`) from v8 to v9. (dnsimple/cli#37)

## 0.8.0 - 2026-05-14

### Changed

- `dnsimple ai` now lists only the required flags by default, which shrinks the output and reduces token usage. Pass `--full` for the complete reference, or run `<command> --help` to discover flags on demand.
- Update the `dnsimple ai` context to remove duplicated command references and to clarify command selection.
- Rename the Go module path from `github.com/dnsimple/dnsimple-cli` to `github.com/dnsimple/cli` after the GitHub repository rename.

## 0.7.0 - 2026-04-24

### Changed

- `dnsimple services apply` and `dnsimple services unapply` now take arguments as `<domain> <service>`, the same order as other domain-scoped commands. This is a breaking change from the previous `<service> <domain>` order. (dnsimple/cli#19)
- Simplify the `vanity-name-servers` command to `vanity-servers`.
- Improve discoverability of the `dnsimple ai` command.
- Remove the singular aliases (`domain`, `zone`, `record`, `template`, `contact`, `cert`) from the plural resource commands to reduce help-output clutter. The `certs` alias for `certificates` is kept as a length-shortening shortcut. (dnsimple/cli#35)

## 0.6.0 - 2026-04-22

### Added

- `dnsimple research status <domain>` checks domain availability through the Domain Research endpoint. The endpoint is designed for high-volume bulk lookups and requires dedicated paid access. (dnsimple/cli#4)
- `dnsimple ai` now includes a "Choosing the Right Command" section. It documents when to use `registrar check` (rate-limited, available on every plan) and when to use `research status` (bulk lookups, paid add-on). (dnsimple/cli#4)

### Fixed

- `dnsimple completion` now shows help text when invoked without a shell argument, instead of the error `Error: accepts 1 arg(s), received 0`. The command is now a parent command with `bash`, `zsh`, `fish`, and `powershell` subcommands. (dnsimple/cli#38)

## 0.5.2 - 2026-04-15

### Added

- Add a native Windows `install.ps1` installer and publish it alongside `install.sh`. (dnsimple/cli#39)

## 0.5.1 - 2026-04-08

### Changed

- The release pipeline now mints a short-lived GitHub App installation token to push Homebrew formula updates to `dnsimple/homebrew-tap`, which replaces the previous personal access token. GitHub signs the commits server-side, and the commits satisfy the tap's verified-signature branch protection rule. (dnsimple/cli#34)

## 0.5.0 - 2026-04-08

### Added

- `auth login` now stores credentials under a named context (kubectl-style); `auth list` shows all stored contexts, `auth switch` selects the active one, and `auth logout --name <name>` removes one. The new `--context <name>` global flag overrides the active context for a single invocation, intended for agents and parallel shells. (dnsimple/cli#28)
- `auth login` masks token input on a TTY when reading interactively. (dnsimple/cli#28)

### Changed

- The credentials file (`credentials.yml`) now stores a list of named contexts instead of a host-keyed map. Files that use the legacy `hosts:` schema migrate automatically and in place on first load. (dnsimple/cli#28)
- `--token`, `--account`, `--sandbox`, and `--context` each fall back independently to environment variables and then to the active stored context. A single override does not require the other values. (dnsimple/cli#28)
- Resource-level destructive commands such as `delete` and `unapply` now prompt for confirmation in interactive terminals. Non-interactive use requires `--yes`. (dnsimple/cli#29)

### Fixed

- `domains delete` now preflights the domain state and requires stronger acknowledgment before it deletes a registered domain. The operation downgrades the domain to `hosted` and permanently loses the registration metadata. (dnsimple/cli#16)
- `--format` now evaluates templates against the underlying resource data for wrapper-backed commands such as `domains get` and `domains list`. Templates like `{{.Name}}` and `{{range .}}{{.Name}}{{end}}` work as expected. (dnsimple/cli#17)
- `auth switch` now validates that the requested account is accessible with the current token, instead of silently storing any value. (dnsimple/cli#25)
- `auth status` now reports the resolved default account (the account commands actually use) instead of the token's bound account from `whoami`. (dnsimple/cli#25)
- `auth status` now reflects the active context's environment when the active context is sandbox and `--sandbox` is not passed on the command line. (dnsimple/cli#28)
- `analytics query` now renders only the columns that match the requested `--groupings`, plus `VOLUME`. It validates the grouping keys (`zone_name`, `date`) before it calls the API and rejects unknown values with a clear error. (dnsimple/cli#21)

## 0.4.0 - 2026-04-03

### Added

- Add an update notification that informs users when a newer CLI version is available. (dnsimple/cli#10)

## 0.3.0 - 2026-04-03

### Added

- Add Homebrew formula publishing via `dnsimple/homebrew-tap`.

## 0.2.0 - 2026-04-03

### Added

- Add an install script for curl-based installation with checksum verification. (dnsimple/cli#9)

## 0.1.0 - 2026-04-02

Initial release.

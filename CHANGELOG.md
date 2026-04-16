# Changelog

This project uses [Semantic Versioning 2.0.0](http://semver.org/), the format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Unreleased

### Added

- `dnsimple research status <domain>` checks domain availability through the Domain Research endpoint, designed for high-volume bulk lookups and requiring dedicated paid access. (dnsimple/dnsimple-cli#4)
- `dnsimple ai` now includes a "Choosing the Right Command" section that documents when to use `registrar check` (rate-limited, available on every plan) versus `research status` (bulk lookups, paid add-on). (dnsimple/dnsimple-cli#4)

### Fixed

- `dnsimple completion` now shows help text when invoked without a shell argument instead of returning `Error: accepts 1 arg(s), received 0`. Restructured as a parent command with `bash`, `zsh`, `fish`, and `powershell` subcommands. (dnsimple/dnsimple-cli#38)

## 0.5.2 - 2026-04-15

### Added

- Add a native Windows `install.ps1` installer and publish it alongside `install.sh`.

## 0.5.1 - 2026-04-08

### Changed

- Release pipeline now mints a short-lived GitHub App installation token to push Homebrew formula updates to `dnsimple/homebrew-tap`, replacing the previous personal access token. Commits are server-side signed by GitHub and satisfy the tap's verified-signature branch protection rule. (dnsimple/dnsimple-cli#34)

## 0.5.0 - 2026-04-08

### Added

- Named authentication contexts: `auth login` now stores credentials under a named context (kubectl-style), `auth list` shows all stored contexts with the active one highlighted, `auth switch` selects the active context by name or account ID (with an interactive picker when called with no arguments), and `auth logout --name <name>` removes a specific context. Multiple production and sandbox contexts can coexist. A new `--context <name>` global flag provides stateless per-invocation overrides: run a command against a stored context without changing the active selection, intended for agents and parallel shells. (dnsimple/dnsimple-cli#28)
- `auth login` masks token input on a TTY when reading interactively. (dnsimple/dnsimple-cli#28)

### Changed

- Credentials file (`credentials.yml`) now stores a list of named contexts instead of a host-keyed map. Existing files using the legacy `hosts:` schema are migrated automatically and in place on first load. (dnsimple/dnsimple-cli#28)
- Field-by-field resolution chain: `--token`, `--account`, `--sandbox`, and `--context` each fall back independently to environment variables and then to the active stored context, so a single override does not require respecifying the others. (dnsimple/dnsimple-cli#28)
- Resource-level destructive commands such as `delete` and `unapply` now prompt for confirmation in interactive terminals and require `--yes` in non-interactive use. (dnsimple/dnsimple-cli#29)

### Fixed

- `domains delete` now preflights the domain state and requires stronger acknowledgment before deleting a registered domain, because the operation downgrades it to `hosted` and permanently loses registration metadata. (dnsimple/dnsimple-cli#16)
- `--format` now evaluates templates against the underlying resource data for wrapper-backed commands such as `domains get` and `domains list`, so templates like `{{.Name}}` and `{{range .}}{{.Name}}{{end}}` work as expected. (dnsimple/dnsimple-cli#17)
- `auth switch` now validates that the requested account is accessible with the current token, instead of silently storing any value. (dnsimple/dnsimple-cli#25)
- `auth status` now reports the resolved default account (the one commands actually use) instead of the token's bound account from `whoami`. (dnsimple/dnsimple-cli#25)
- `auth status` now reflects the active context's environment when the active context is sandbox without `--sandbox` being passed on the command line. (dnsimple/dnsimple-cli#28)
- `analytics query` now renders only the columns that match the requested `--groupings` (plus `VOLUME`), validates supported grouping keys (`zone_name`, `date`) before calling the API, and rejects unknown values with a clear error. (dnsimple/dnsimple-cli#21)

## 0.4.0 - 2026-04-03

### Added

- Add update notification that informs users when a newer CLI version is available. (dnsimple/dnsimple-cli#10)

## 0.3.0 - 2026-04-03

### Added

- Add Homebrew formula publishing via `dnsimple/homebrew-tap`.

## 0.2.0 - 2026-04-03

### Added

- Install script for curl-based installation with checksum verification. (dnsimple/dnsimple-cli#9)

## 0.1.0 - 2026-04-02

Initial release.

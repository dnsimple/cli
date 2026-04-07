# Changelog

This project uses [Semantic Versioning 2.0.0](http://semver.org/), the format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## main

### Added

- Named authentication contexts: `auth login` now stores credentials under a named context (kubectl-style), `auth list` shows all stored contexts with the active one highlighted, `auth switch` selects the active context by name or account ID (with an interactive picker when called with no arguments), and `auth logout --name <name>` removes a specific context. Multiple production and sandbox contexts can coexist. A new `--context <name>` global flag provides stateless per-invocation overrides: run a command against a stored context without changing the active selection, intended for agents and parallel shells. (dnsimple/dnsimple-cli#28)
- `auth login` masks token input on a TTY when reading interactively. (dnsimple/dnsimple-cli#28)

### Changed

- Credentials file (`credentials.yml`) now stores a list of named contexts instead of a host-keyed map. Existing files using the legacy `hosts:` schema are migrated automatically and in place on first load. (dnsimple/dnsimple-cli#28)
- Field-by-field resolution chain: `--token`, `--account`, `--sandbox`, and `--context` each fall back independently to environment variables and then to the active stored context, so a single override does not require respecifying the others. (dnsimple/dnsimple-cli#28)
- Resource-level destructive commands such as `delete` and `unapply` now prompt for confirmation in interactive terminals and require `--yes` in non-interactive use. (dnsimple/dnsimple-cli#29)

### Fixed

- `--format` now evaluates templates against the underlying resource data for wrapper-backed commands such as `domains get` and `domains list`, so templates like `{{.Name}}` and `{{range .}}{{.Name}}{{end}}` work as expected. (dnsimple/dnsimple-cli#17)
- `auth switch` now validates that the requested account is accessible with the current token, instead of silently storing any value. (dnsimple/dnsimple-cli#25)
- `auth status` now reports the resolved default account (the one commands actually use) instead of the token's bound account from `whoami`. (dnsimple/dnsimple-cli#25)
- `auth status` now reflects the active context's environment when the active context is sandbox without `--sandbox` being passed on the command line. (dnsimple/dnsimple-cli#28)

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

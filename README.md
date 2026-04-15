# DNSimple CLI

A command-line interface for the [DNSimple API v2](https://developer.dnsimple.com/v2/).

## Requirements

- Go 1.25+
- An activated DNSimple account

## Installation

### Windows

#### Using the PowerShell install script

```powershell
powershell -ExecutionPolicy Bypass -Command "irm https://cli-beta.dnsimple.com/install.ps1 | iex"
```

The installer downloads release artifacts from the public `dnsimple/homebrew-tap` release mirror.

#### Manual ZIP install

1. Download the latest `dnsimple_<version>_windows_amd64.zip` or `dnsimple_<version>_windows_arm64.zip` from the [public release mirror](https://github.com/dnsimple/homebrew-tap/releases/latest) page.
2. Extract `dnsimple.exe`.
3. Move `dnsimple.exe` to a directory on your `PATH`, such as `%USERPROFILE%\bin`.
4. Open a new terminal window and verify the installation:

```powershell
dnsimple version
```

> [!NOTE]
> Homebrew is supported on macOS, Linux, and WSL. The `install.sh` script can also be used from WSL or Unix-like shells such as Git Bash, but it is not a native `powershell` or `cmd.exe` installer.

### macOS and Linux

#### Using the install script

```shell
curl -fsSL http://cli-beta.dnsimple.com/install.sh | sh
```

The install URL will eventually move to `https://dnsimple.com/install.sh`.

The install script downloads release artifacts from the public `dnsimple/homebrew-tap` release mirror.

#### Using Homebrew

```shell
brew install dnsimple/tap/dnsimple
```

### Cross-platform developer install

#### Using Go

```shell
go install github.com/dnsimple/dnsimple-cli/cmd/dnsimple@latest
```

#### From source

```shell
git clone https://github.com/dnsimple/dnsimple-cli.git
cd dnsimple-cli
make build
```

The binary will be available in the `bin/` directory.

## Usage

```shell
dnsimple [command] [flags]
```

### Authentication

The CLI supports two authentication modes that can be combined freely.

> [!NOTE]
> The CLI currently supports API token authentication only, including both classic and scoped API tokens. OAuth support may be considered in the future, but it is not currently on the roadmap.

#### Stateful: stored contexts

Authenticate once and the CLI remembers a named *context* (token, account, environment) on disk. Multiple contexts can coexist and you select one as active:

```shell
# Log in to production and store a context
dnsimple auth login

# Log in to sandbox alongside it
dnsimple auth login --sandbox

# List stored contexts (active is marked with *)
dnsimple auth list

# Switch the active context (by name or by account ID)
dnsimple auth switch sandbox

# Inspect the active context
dnsimple auth status

# Remove a stored context
dnsimple auth logout --name sandbox
```

The active context is used by every command unless overridden. Pass `--name` to `auth login` to choose a custom context name; otherwise it is derived from the environment (`production`, `sandbox`) with the account ID appended on collision.

#### Stateless: per-invocation overrides

For agents, scripts, and parallel shells where mutating shared on-disk state is undesirable, override the active context for a single command:

```shell
# Use a stored context by name without switching the active one
dnsimple --context sandbox zones list

# Override individual fields (token from env, sandbox environment)
DNSIMPLE_TOKEN=$TOK dnsimple --sandbox zones list

# Fully stateless invocation
dnsimple --token $TOK --account 1010 --sandbox zones list
```

The override chain is field-by-field: each of `--token`, `--account`, `--sandbox`, and `--context` falls back to the matching environment variable and then to the active stored context. This means a script can supply only the parts that differ from the active context.

### Destructive Commands

Resource-level destructive commands such as `delete` and `unapply` now prompt for confirmation in interactive terminals. In scripts, CI, or other non-interactive use, pass `--yes` to confirm explicitly and proceed without prompting.

Registered domains are a higher-risk case: `dnsimple domains delete` first checks the domain state, and deleting a registered domain requires an extra acknowledgment because the operation downgrades the domain to `hosted` and permanently loses registration metadata. In non-interactive use, pass both `--yes` and `--confirm-registered-domain`.

### Example Flow

A common workflow looks like this:

1. Authenticate
2. Add a domain to your account
3. Activate DNS for the zone
4. Create, update, and delete records

Example:

```shell
# 1. Authenticate and verify the current identity
dnsimple auth login
dnsimple auth status

# 2. Add a domain to your account
dnsimple domains create example.com

# 3. Activate DNS for the zone
dnsimple zones activate example.com

# 4a. Create a record
dnsimple records create example.com \
  --type A \
  --name www \
  --content 192.0.2.10 \
  --ttl 3600

# 4b. List records so you can grab the record ID
dnsimple records list example.com

# 4c. Update the record
dnsimple records update example.com 12345 \
  --content 192.0.2.20 \
  --ttl 600

# 4d. Delete the record
dnsimple records delete example.com 12345
```

`example.com` is both the domain name and the zone name. After creating the domain, use the same value with the `zones` and `records` commands.

### Output Formats

Commands default to a human-friendly table. Use `--json` for structured output or `--format` for custom Go templates.

`--json` preserves the CLI wrapper shape, including top-level `data` and optional `pagination` keys. `--format` evaluates the template against the unwrapped resource itself:

```shell
# Single-resource commands expose the resource fields directly
dnsimple domains get example.com --format '{{.Name}}'

# List commands expose the list of resources directly
dnsimple domains list --format '{{range .}}{{.Name}}{{printf "\n"}}{{end}}'
```

To discover available template fields, inspect the corresponding `--json` output and use the fields inside `data` as the underlying resource model.

### Sandbox Environment

We highly recommend testing against our [sandbox environment](https://developer.dnsimple.com/sandbox/) before using our production environment. This will allow you to avoid real purchases, live charges on your credit card, and reduce the chance of your running up against rate limits.

To use the sandbox environment, either store a sandbox context:

```shell
dnsimple auth login --sandbox
```

Or pass `--sandbox` per invocation:

```shell
dnsimple --sandbox zones list
```

You will need a token created in the sandbox environment. Production tokens will *not* work in the sandbox environment.

## Documentation

- [DNSimple API documentation](https://developer.dnsimple.com)
- [DNSimple support documentation](https://support.dnsimple.com)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

Copyright (c) 2026 DNSimple Corporation. This is Free Software distributed under the [MIT License](LICENSE.txt).

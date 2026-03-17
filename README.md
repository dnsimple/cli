# DNSimple CLI

A command-line interface for the [DNSimple API v2](https://developer.dnsimple.com/v2/).

## Requirements

- Go 1.25+
- An activated DNSimple account

## Installation

### Using Go

```shell
go install github.com/dnsimple/dnsimple-cli/cmd/dnsimple@latest
```

### From source

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

Set your DNSimple API token:

```shell
export DNSIMPLE_TOKEN=your-token
```

Or pass it directly:

```shell
dnsimple --token your-token [command]
```

### Typical Flow

A common workflow looks like this:

1. Authenticate
2. Add a domain to your account
3. Activate DNS for the zone
4. Create, update, and delete records

Example:

```shell
# 1. Authenticate and verify the current identity
export DNSIMPLE_TOKEN=your-token
printf '%s\n' "$DNSIMPLE_TOKEN" | dnsimple auth login --with-token
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

### Sandbox Environment

We highly recommend testing against our [sandbox environment](https://developer.dnsimple.com/sandbox/) before using our production environment. This will allow you to avoid real purchases, live charges on your credit card, and reduce the chance of your running up against rate limits.

To use the sandbox environment:

```shell
export DNSIMPLE_BASE_URL=https://api.sandbox.dnsimple.com
```

You will need to ensure that you are using an access token created in the sandbox environment. Production tokens will *not* work in the sandbox environment.

## Documentation

- [DNSimple API documentation](https://developer.dnsimple.com)
- [DNSimple support documentation](https://support.dnsimple.com)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

Copyright (c) 2026 DNSimple Corporation. This is Free Software distributed under the [MIT License](LICENSE.txt).

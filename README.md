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

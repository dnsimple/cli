package main

import (
	"os"

	"github.com/dnsimple/dnsimple-cli/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	code := cli.Execute(version, os.Args[1:])
	os.Exit(code)
}

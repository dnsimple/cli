package client

import (
	"context"

	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
)

const cliVersion = "0.1.0"

// NewClient creates a new DNSimple API client configured with the given token and config.
func NewClient(cfg *config.Config, token string) *dnsimple.Client {
	httpClient := dnsimple.StaticTokenHTTPClient(context.Background(), token)
	client := dnsimple.NewClient(httpClient)
	client.SetUserAgent("dnsimple-cli/" + cliVersion)
	client.BaseURL = cfg.BaseURL
	return client
}

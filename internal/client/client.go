package client

import (
	"context"
	"fmt"

	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
)

// NewClient creates a new DNSimple API client configured with the given token and config.
func NewClient(cfg *config.Config, token string, version string) *dnsimple.Client {
	httpClient := dnsimple.StaticTokenHTTPClient(context.Background(), token)
	client := dnsimple.NewClient(httpClient)
	client.SetUserAgent(fmt.Sprintf("dnsimple-cli/%s", version))
	client.BaseURL = cfg.BaseURL
	return client
}

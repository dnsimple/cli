package client

import (
	"context"
	"fmt"

	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
)

// Options configures a DNSimple API client.
type Options struct {
	// BaseURL is the API URL the client should target.
	BaseURL string

	// Token is the bearer token sent with each request.
	Token string

	// Version is used to build the User-Agent string.
	Version string

	// Debug enables verbose debug logging on the underlying client.
	Debug bool
}

// New creates a new DNSimple API client from the given options.
func New(opts Options) *dnsimple.Client {
	httpClient := dnsimple.StaticTokenHTTPClient(context.Background(), opts.Token)
	c := dnsimple.NewClient(httpClient)
	c.SetUserAgent(fmt.Sprintf("dnsimple-cli/%s", opts.Version))
	c.BaseURL = opts.BaseURL
	if opts.Debug {
		c.Debug = true
	}
	return c
}

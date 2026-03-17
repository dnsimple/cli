package client

import (
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewClientUsesConfigAndCLIUserAgent(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.sandbox.dnsimple.com"}

	client := NewClient(cfg, "token", "1.2.3")

	assert.Equal(t, cfg.BaseURL, client.BaseURL)
	assert.Equal(t, "dnsimple-cli/1.2.3", client.UserAgent)
}

func TestNewClientPreservesDevVersionInUserAgent(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.sandbox.dnsimple.com"}

	client := NewClient(cfg, "token", "dev")

	assert.True(t, strings.HasPrefix(client.UserAgent, "dnsimple-cli/"), "UserAgent = %q, want dnsimple-cli/*", client.UserAgent)
	assert.Equal(t, "dnsimple-cli/dev", client.UserAgent)
}

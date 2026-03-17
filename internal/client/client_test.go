package client

import (
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewClientUsesConfigAndCLIUserAgent(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.sandbox.dnsimple.com"}

	client := NewClient(cfg, "token")

	assert.Equal(t, cfg.BaseURL, client.BaseURL)
	assert.True(t, strings.HasPrefix(client.UserAgent, "dnsimple-cli/"), "UserAgent = %q, want dnsimple-cli/*", client.UserAgent)
}

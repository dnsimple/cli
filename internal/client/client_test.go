package client

import (
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/config"
)

func TestNewClientUsesConfigAndCLIUserAgent(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.sandbox.dnsimple.com"}

	client := NewClient(cfg, "token")

	if client.BaseURL != cfg.BaseURL {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL, cfg.BaseURL)
	}
	if !strings.HasPrefix(client.UserAgent, "dnsimple-cli/") {
		t.Fatalf("UserAgent = %q, want dnsimple-cli/*", client.UserAgent)
	}
}

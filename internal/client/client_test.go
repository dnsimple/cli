package client

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsesBaseURLAndCLIUserAgent(t *testing.T) {
	c := New(Options{
		BaseURL: "https://api.sandbox.dnsimple.com",
		Token:   "token",
		Version: "1.2.3",
	})

	assert.Equal(t, "https://api.sandbox.dnsimple.com", c.BaseURL)
	assert.Equal(t, "dnsimple-cli/1.2.3", c.UserAgent)
	assert.False(t, c.Debug)
}

func TestNewPreservesDevVersionInUserAgent(t *testing.T) {
	c := New(Options{
		BaseURL: "https://api.sandbox.dnsimple.com",
		Token:   "token",
		Version: "dev",
	})

	assert.True(t, strings.HasPrefix(c.UserAgent, "dnsimple-cli/"), "UserAgent = %q, want dnsimple-cli/*", c.UserAgent)
	assert.Equal(t, "dnsimple-cli/dev", c.UserAgent)
}

func TestNewEnablesDebugWhenRequested(t *testing.T) {
	c := New(Options{
		BaseURL: "https://api.dnsimple.com",
		Token:   "token",
		Version: "test",
		Debug:   true,
	})

	assert.True(t, c.Debug)
}

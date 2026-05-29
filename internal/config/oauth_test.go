package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthTokenURLAppendsPathToBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"production", "https://api.dnsimple.com", "https://api.dnsimple.com/v2/oauth/access_token"},
		{"sandbox", "https://api.sandbox.dnsimple.com", "https://api.sandbox.dnsimple.com/v2/oauth/access_token"},
		{"local test server", "http://127.0.0.1:54321", "http://127.0.0.1:54321/v2/oauth/access_token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, OAuthTokenURL(tc.baseURL))
		})
	}
}

func TestAuthorizeURLPicksHostForEnvironment(t *testing.T) {
	assert.Equal(t, "https://dnsimple.com/oauth/authorize", AuthorizeURL(false))
	assert.Equal(t, "https://sandbox.dnsimple.com/oauth/authorize", AuthorizeURL(true))
}

func TestOAuthClientIDFallsBackToEmbeddedConstants(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID", "")
	assert.Equal(t, oauthClientIDProduction, OAuthClientID(false))
	assert.Equal(t, oauthClientIDSandbox, OAuthClientID(true))
}

func TestOAuthClientIDEnvOverrideTakesPrecedence(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID", "override-id")
	assert.Equal(t, "override-id", OAuthClientID(false))
	assert.Equal(t, "override-id", OAuthClientID(true))
}

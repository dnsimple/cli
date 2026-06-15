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
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION", "")
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_SANDBOX", "")
	assert.Equal(t, oauthClientIDProduction, OAuthClientID(false))
	assert.Equal(t, oauthClientIDSandbox, OAuthClientID(true))
}

func TestOAuthClientIDPerEnvironmentOverridesAreScoped(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION", "prod-id")
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_SANDBOX", "sbx-id")
	assert.Equal(t, "prod-id", OAuthClientID(false))
	assert.Equal(t, "sbx-id", OAuthClientID(true))
}

// TestOAuthClientIDDoesNotCrossEnvironments guards the split: setting one
// env var must not leak into the other environment. A developer
// alternating between prod and sandbox in one shell session would
// otherwise hit dnsimple.com with a sandbox-only ID and see
// invalid_client.
func TestOAuthClientIDDoesNotCrossEnvironments(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_SANDBOX", "sbx-only")
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION", "")
	assert.Equal(t, oauthClientIDProduction, OAuthClientID(false))
	assert.Equal(t, "sbx-only", OAuthClientID(true))
}

func TestOAuthClientIDTrimsWhitespaceFromOverride(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION", "  prod-id  \n")
	assert.Equal(t, "prod-id", OAuthClientID(false))
}

// TestOAuthClientIDIgnoresWhitespaceOnlyOverride guards the case where an
// operator's command-substitution sets the var to whitespace only -- the
// flow must treat it as unset and fall back to the embedded constant,
// not propagate the spaces into client_id.
func TestOAuthClientIDIgnoresWhitespaceOnlyOverride(t *testing.T) {
	t.Setenv("DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION", "   \n  ")
	assert.Equal(t, oauthClientIDProduction, OAuthClientID(false))
}

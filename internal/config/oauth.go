package config

import (
	"os"
	"strings"
)

// OAuth client identifiers for the first-party DNSimple CLI application, one
// per environment. Public OAuth client identifiers are not secrets, so
// embedding them in the binary is standard practice (gh, gcloud, stripe all
// do this). An empty value makes OAuthClientID return "", so `auth login`
// falls back to the paste prompt instead of starting the browser flow.
const (
	oauthClientIDProduction = "902c56782e81c768"
	oauthClientIDSandbox    = "46bbb707b411f835"
)

// Per-environment overrides consulted before the embedded constants.
// Useful when developing against a local dnsimple-app checkout where the
// CLI app has been bootstrapped with a different ID, or when alternating
// between sandbox and production in one shell session: a single shared
// override would route the wrong client ID into one of the two flows and
// produce an opaque `invalid_client` error.
const (
	oauthClientIDEnvVarProduction = "DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION"
	oauthClientIDEnvVarSandbox    = "DNSIMPLE_OAUTH_CLIENT_ID_SANDBOX"
)

// OAuthClientID returns the OAuth client identifier to use for the given
// environment. The matching environment variable
// (DNSIMPLE_OAUTH_CLIENT_ID_PRODUCTION or _SANDBOX) takes precedence over
// the embedded constant. Leading/trailing whitespace on the env value is
// stripped before the empty check so a stray newline from a
// command-substitution export (e.g. `export ...=$(<file)`) does not
// silently propagate into client_id and yield a cryptic browser error.
//
// An empty return value means "OAuth is not provisioned for this build /
// environment yet": callers should fall back to the manual token paste flow.
func OAuthClientID(sandbox bool) string {
	envvar := oauthClientIDEnvVarProduction
	embedded := oauthClientIDProduction
	if sandbox {
		envvar = oauthClientIDEnvVarSandbox
		embedded = oauthClientIDSandbox
	}
	if v := strings.TrimSpace(os.Getenv(envvar)); v != "" {
		return v
	}
	return embedded
}

// oauthAuthorizeURLEnvVar overrides the authorize endpoint. It exists for
// local development against a dnsimple-app checkout, where the authorize
// flow runs on http://localhost:3000 rather than the production or sandbox
// hosts. Unset in normal use.
const oauthAuthorizeURLEnvVar = "DNSIMPLE_OAUTH_AUTHORIZE_URL"

// AuthorizeURL returns the OAuth authorize endpoint for the given
// environment. The authorize endpoint lives on the application host
// (dnsimple.com / sandbox.dnsimple.com), not the API host, which is why
// it has its own helper rather than being derived from BaseURLForHost.
//
// DNSIMPLE_OAUTH_AUTHORIZE_URL, when set, takes precedence over both
// environments for local development. Leading/trailing whitespace is
// stripped before the empty check, matching OAuthClientID.
func AuthorizeURL(sandbox bool) string {
	if v := strings.TrimSpace(os.Getenv(oauthAuthorizeURLEnvVar)); v != "" {
		return v
	}
	if sandbox {
		return "https://sandbox.dnsimple.com/oauth/authorize"
	}
	return "https://dnsimple.com/oauth/authorize"
}

// OAuthTokenURL returns the OAuth token endpoint relative to the given
// API base URL. The base URL is passed in directly (rather than
// re-derived from a `sandbox` bool) so it stays in sync with whatever
// the rest of the API client is using: if a future BaseURL override
// ships (env var, enterprise host, integration-test seam), OAuth
// follows it automatically. Pass cfg.BaseURL.
func OAuthTokenURL(baseURL string) string {
	return baseURL + "/v2/oauth/access_token"
}

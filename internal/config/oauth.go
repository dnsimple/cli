package config

import "os"

// OAuth client identifiers for the first-party DNSimple CLI application.
//
// These values are populated by a follow-up rollout PR once the dnsimple-app
// side bootstraps the public OAuth application in each environment (see
// `docs/oauth.md` in dnsimple-app). Until they are populated, OAuthClientID
// returns the empty string and `auth login` falls back to the paste prompt
// with an explanatory message instead of starting the browser flow.
//
// Client identifiers for public OAuth clients are not secrets, so embedding
// them in the binary is standard practice (gh, gcloud, stripe all do this).
const (
	oauthClientIDProduction = ""
	oauthClientIDSandbox    = ""
)

// oauthClientIDEnvVar is the per-invocation override consulted before the
// embedded constants. Useful when developing against a local dnsimple-app
// checkout where the CLI app has been bootstrapped with a different ID.
const oauthClientIDEnvVar = "DNSIMPLE_OAUTH_CLIENT_ID"

// OAuthClientID returns the OAuth client identifier to use for the given
// environment. The DNSIMPLE_OAUTH_CLIENT_ID environment variable takes
// precedence over the embedded constants when set.
//
// An empty return value means "OAuth is not provisioned for this build /
// environment yet": callers should fall back to the manual token paste flow.
func OAuthClientID(sandbox bool) string {
	if v := os.Getenv(oauthClientIDEnvVar); v != "" {
		return v
	}
	if sandbox {
		return oauthClientIDSandbox
	}
	return oauthClientIDProduction
}

// AuthorizeURL returns the OAuth authorize endpoint for the given
// environment. The authorize endpoint lives on the application host
// (dnsimple.com / sandbox.dnsimple.com), not the API host, which is why
// it has its own helper rather than being derived from BaseURLForHost.
func AuthorizeURL(sandbox bool) string {
	if sandbox {
		return "https://sandbox.dnsimple.com/oauth/authorize"
	}
	return "https://dnsimple.com/oauth/authorize"
}

// OAuthTokenURL returns the OAuth token endpoint for the given environment.
// The token endpoint shares the API host (api.dnsimple.com), so it is
// derived from BaseURLForHost.
func OAuthTokenURL(sandbox bool) string {
	return BaseURLForHost(HostForSandbox(sandbox)) + "/v2/oauth/access_token"
}

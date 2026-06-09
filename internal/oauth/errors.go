// Package oauth implements the interactive OAuth 2.0 Authorization Code
// flow with PKCE (RFC 7636) and a loopback redirect (RFC 8252 §7.3) for
// `dnsimple auth login`.
//
// The flow is:
//
//  1. Generate a PKCE verifier and SHA-256 challenge.
//  2. Generate a random state.
//  3. Bind 127.0.0.1 on an OS-assigned port and listen for the callback.
//  4. Open the user's browser to the authorize URL.
//  5. Receive ?code=…&state=… on the loopback, validate state, exchange
//     code + verifier at the token endpoint for an access token.
//
// The token endpoint accepts JSON and does not require a client_secret
// for public clients (PKCE replaces the secret).
package oauth

import "errors"

// ErrNotProvisioned is returned by Login when the caller did not supply a
// client ID. Callers (typically the CLI `auth login` command) treat this as
// "OAuth is not enabled for this build", and fall back to the manual token
// paste prompt.
var ErrNotProvisioned = errors.New("oauth client id not configured")

// ErrStateMismatch is returned when the state parameter on the callback
// does not match the one generated for the request. It indicates either
// a stale browser tab or an attempt to inject a forged callback.
var ErrStateMismatch = errors.New("oauth: state mismatch on callback")

// AuthError carries an RFC 6749 §4.1.2.1 or §5.2 error response from the
// authorization server. The CLI surfaces both Code and Description so the
// user sees a useful message ("access_denied: user cancelled the request").
//
// Construct via newAuthError so attacker-controlled bytes are stripped
// before they can reach the terminal via Error().
type AuthError struct {
	Code        string
	Description string
}

// newAuthError builds an AuthError, stripping C0 control bytes and DEL
// from both fields. The values flow from either an OAuth-server JSON
// body or a query-string parameter on the loopback redirect, and either
// source can carry attacker-controlled content -- left raw they would
// emit ANSI escape sequences when the CLI prints err.Error() to the
// terminal. The HTML error page goes through htmlEscape; this is the
// terminal-side equivalent.
func newAuthError(code, description string) *AuthError {
	return &AuthError{
		Code:        sanitizeForTerminal(code),
		Description: sanitizeForTerminal(description),
	}
}

func (e *AuthError) Error() string {
	if e.Description != "" {
		return e.Code + ": " + e.Description
	}
	return e.Code
}

// sanitizeForTerminal drops every C0 control byte (0x00-0x1F) and DEL
// (0x7F) from s. UTF-8 continuation and lead bytes are all >= 0x80, so
// non-ASCII text passes through unchanged.
func sanitizeForTerminal(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 || b == 0x7f {
			continue
		}
		out = append(out, b)
	}
	return string(out)
}

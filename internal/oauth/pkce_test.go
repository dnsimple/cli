package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestChallengeRFC7636Vector pins the S256 derivation against the published
// vector from RFC 7636 §4.6. Any drift here means the implementation no
// longer interoperates with the server's PKCE check.
func TestChallengeRFC7636Vector(t *testing.T) {
	const (
		verifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	)
	got := challengeFor(verifier)
	assert.Equal(t, challenge, got)
}

// challengeFor is a one-line trampoline so the test name reads naturally
// alongside the unexported `challenge` function.
func challengeFor(v string) string { return challenge(v) }

func TestNewVerifierIsBase64URLAndLongEnough(t *testing.T) {
	v, err := newVerifier()
	if !assert.NoError(t, err) {
		return
	}
	// 32 bytes base64url-no-padding-encoded → 43 characters.
	assert.Len(t, v, 43)
	// Only RFC 7636 §4.1 unreserved chars: ALPHA / DIGIT / "-" / "." / "_" / "~".
	// base64url uses ALPHA / DIGIT / "-" / "_"; that's a strict subset.
	for _, r := range v {
		isAllowed := (r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_'
		assert.Truef(t, isAllowed, "verifier contains disallowed char %q", r)
	}
}

func TestNewVerifierIsRandom(t *testing.T) {
	a, err := newVerifier()
	if !assert.NoError(t, err) {
		return
	}
	b, err := newVerifier()
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEqual(t, a, b)
}

func TestNewStateIsRandomAndBase64URL(t *testing.T) {
	a, err := newState()
	if !assert.NoError(t, err) {
		return
	}
	b, err := newState()
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEqual(t, a, b)
	assert.Len(t, a, 43)
}

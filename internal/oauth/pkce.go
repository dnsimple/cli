package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// verifierBytes is the entropy size for the PKCE code verifier. 32 random
// bytes encode to 43 unreserved base64url characters, which sits at the low
// end of RFC 7636 §4.1's [43, 128] range and gives ~256 bits of entropy.
const verifierBytes = 32

// newVerifier returns a fresh PKCE code verifier per RFC 7636 §4.1: a
// base64url-encoded random byte string with no padding.
func newVerifier() (string, error) {
	buf := make([]byte, verifierBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes for verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// challenge returns the S256 code challenge for the given verifier per RFC
// 7636 §4.2: base64url(SHA-256(verifier)) with no padding.
func challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// stateBytes is the entropy size for the OAuth `state` parameter. 32 bytes
// (~256 bits) is well above the typical recommendation for CSRF protection.
const stateBytes = 32

// newState returns a fresh OAuth state token: a base64url-encoded random
// byte string with no padding. The CLI generates one per login attempt and
// validates it on the callback to detect forged or stale redirects.
func newState() (string, error) {
	buf := make([]byte, stateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes for state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

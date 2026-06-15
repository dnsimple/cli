package oauth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeForTerminalStripsC0AndDEL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ansi color", "\x1b[31mred\x1b[0m", "[31mred[0m"},
		{"bel and cr", "alert\x07\rback", "alertback"},
		{"del", "abc\x7fdef", "abcdef"},
		{"plain ascii passes through", "no escapes here", "no escapes here"},
		{"utf-8 multibyte preserved", "café — résumé", "café — résumé"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitizeForTerminal(tc.in))
		})
	}
}

// TestNewAuthErrorStripsControlBytes guards the boundary where attacker-
// controlled query-string / JSON content lands in AuthError. Error() must
// not expose any byte that a terminal would interpret as a control or
// escape sequence.
func TestNewAuthErrorStripsControlBytes(t *testing.T) {
	err := newAuthError("\x1b[31mFAKE\x07", "evil\x1b[2Jcleared")
	assert.Equal(t, "[31mFAKE", err.Code)
	assert.Equal(t, "evil[2Jcleared", err.Description)

	msg := err.Error()
	for _, b := range []byte(msg) {
		assert.Falsef(t, b < 0x20 || b == 0x7f, "AuthError.Error() leaked control byte 0x%02x", b)
	}
	// And the human-readable join still works.
	assert.True(t, strings.Contains(msg, "FAKE"))
	assert.True(t, strings.Contains(msg, "cleared"))
}

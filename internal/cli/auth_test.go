package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestPromptForAccountSelectionReturnsErrorOnEOF(t *testing.T) {
	accountID, err := promptForAccountSelection(strings.NewReader(""), io.Discard, []dnsimple.Account{
		{ID: 101, Email: "one@example.com"},
		{ID: 202, Email: "two@example.com"},
	})

	assert.EqualError(t, err, "no account selected")
	assert.Empty(t, accountID)
}

func TestPromptForAccountSelectionReturnsSelectedAccountID(t *testing.T) {
	accountID, err := promptForAccountSelection(strings.NewReader("2\n"), io.Discard, []dnsimple.Account{
		{ID: 101, Email: "one@example.com"},
		{ID: 202, Email: "two@example.com"},
	})

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "202", accountID)
}

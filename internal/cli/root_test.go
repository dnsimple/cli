package cli

import (
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestExecuteVersionReturnsSuccess(t *testing.T) {
	code := Execute("1.2.3", []string{"--version"})
	assert.Equal(t, cmdutil.ExitOK, code)
}

func TestExecuteUnknownCommandReturnsError(t *testing.T) {
	code := Execute("1.2.3", []string{"definitely-not-a-command"})
	assert.Equal(t, cmdutil.ExitError, code)
}

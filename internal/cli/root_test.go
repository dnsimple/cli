package cli

import (
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
)

func TestExecuteVersionReturnsSuccess(t *testing.T) {
	code := Execute("1.2.3", []string{"--version"})
	if code != cmdutil.ExitOK {
		t.Fatalf("Execute() = %d, want %d", code, cmdutil.ExitOK)
	}
}

func TestExecuteUnknownCommandReturnsError(t *testing.T) {
	code := Execute("1.2.3", []string{"definitely-not-a-command"})
	if code != cmdutil.ExitError {
		t.Fatalf("Execute() = %d, want %d", code, cmdutil.ExitError)
	}
}

package cli

import (
	"bytes"
	"strings"
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

func TestHelpUsesDeclaredLocalFlagOrder(t *testing.T) {
	analyticsHelp := commandHelpOutput(t, "analytics", "query")
	assertContainsInOrder(t, analyticsHelp,
		"--start-date",
		"--end-date",
		"--groupings",
		"--sort",
		"--page",
		"--per-page",
	)

	domainsHelp := commandHelpOutput(t, "domains", "list")
	assertContainsInOrder(t, domainsHelp,
		"--name-like",
		"--all",
		"--sort",
		"--page",
		"--per-page",
	)
}

func commandHelpOutput(t *testing.T, args ...string) string {
	t.Helper()

	f := cmdutil.NewFactory("test")
	root := buildRootCmd(f)

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs(append(args, "--help"))

	err := root.Execute()
	assert.NoError(t, err)

	return stdout.String()
}

func assertContainsInOrder(t *testing.T, output string, values ...string) {
	t.Helper()

	previous := -1
	for _, value := range values {
		index := strings.Index(output, value)
		if assert.NotEqualf(t, -1, index, "expected %q in output", value) {
			assert.Greaterf(t, index, previous, "expected %q after previous value", value)
			previous = index
		}
	}
}

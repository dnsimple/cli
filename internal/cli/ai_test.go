package cli

import (
	"bytes"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestAICommandReturnsSuccess(t *testing.T) {
	code := Execute("1.2.3", []string{"ai"})
	assert.Equal(t, cmdutil.ExitOK, code)
}

func runAICommand(t *testing.T, extraArgs ...string) string {
	t.Helper()

	f := cmdutil.NewFactory("test")
	root := buildRootCmd(f)
	root.AddCommand(newAICmd())

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs(append([]string{"ai"}, extraArgs...))

	err := root.Execute()
	assert.NoError(t, err)

	return stdout.String()
}

func TestAICommandOutputContainsPreamble(t *testing.T) {
	output := runAICommand(t)

	assert.Contains(t, output, "# DNSimple CLI — AI Context")
	assert.Contains(t, output, "## Authentication")
	assert.Contains(t, output, "## Global Flags")
	assert.Contains(t, output, "## Output Formats")
	assert.Contains(t, output, "## Commands")
	assert.Contains(t, output, "## Common Workflows")
	assert.Contains(t, output, "## Tips")
}

func TestAICommandOutputContainsDynamicCommands(t *testing.T) {
	output := runAICommand(t)

	assert.Contains(t, output, "dnsimple zones records create")
	assert.Contains(t, output, "dnsimple registrar register")
	assert.Contains(t, output, "dnsimple domains list")
	assert.Contains(t, output, "dnsimple certificates")
	assert.Contains(t, output, "dnsimple whoami")
}

func TestAICommandSlimOutputListsRequiredFlags(t *testing.T) {
	output := runAICommand(t)

	assert.Contains(t, output, "Required flags:")
	assert.NotContains(t, output, "**(required)**")
}

func TestAICommandFullOutputMarksRequiredFlags(t *testing.T) {
	output := runAICommand(t, "--full")

	assert.Contains(t, output, "**(required)**")
	assert.Contains(t, output, "Flags:")
}

func TestAICommandFullOutputIsLongerThanSlim(t *testing.T) {
	slim := runAICommand(t)
	full := runAICommand(t, "--full")

	assert.Greater(t, len(full), len(slim))
}

func TestAICommandExcludesItself(t *testing.T) {
	output := runAICommand(t)

	assert.NotContains(t, output, "#### `dnsimple ai`")
	assert.NotContains(t, output, "#### `dnsimple completion")
}

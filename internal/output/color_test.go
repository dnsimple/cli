package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorEnabled(t *testing.T) {
	t.Run("non-terminal writer", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		assert.False(t, ColorEnabled(&bytes.Buffer{}, false))
	})

	t.Run("no-color flag", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		assert.False(t, ColorEnabled(os.Stdout, true))
	})

	t.Run("NO_COLOR env var", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		assert.False(t, ColorEnabled(os.Stdout, false))
	})
}

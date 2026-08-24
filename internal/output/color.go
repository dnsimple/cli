package output

import (
	"io"
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// IsTerminal reports whether w is a terminal. A writer that is not an *os.File
// never is, so redirected output and captured test output stay plain.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// ColorEnabled reports whether w accepts colored output. The noColor argument
// carries the --no-color flag.
func ColorEnabled(w io.Writer, noColor bool) bool {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTerminal(w)
}

// NewColor builds a color that is explicitly on or off. fatih/color decides
// globally from its own probe of the standard output stream, so each color the
// CLI builds must override that decision for the writer it goes to.
func NewColor(enabled bool, attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if enabled {
		c.EnableColor()
	} else {
		c.DisableColor()
	}
	return c
}

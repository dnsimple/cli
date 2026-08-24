package output

import (
	"io"
	"os"

	"golang.org/x/term"
)

// ColorEnabled reports whether w accepts colored output. The noColor argument
// carries the --no-color flag. A writer that is not a terminal never gets color,
// so redirected output and captured test output stay plain.
func ColorEnabled(w io.Writer, noColor bool) bool {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

//go:build unix

package output

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func detectTerminalWidth(w io.Writer) (int, bool) {
	file, ok := w.(*os.File)
	if !ok {
		return 0, false
	}

	ws, err := unix.IoctlGetWinsize(int(file.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws == nil || ws.Col == 0 {
		return 0, false
	}

	return int(ws.Col), true
}

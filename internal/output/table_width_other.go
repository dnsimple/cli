//go:build !unix

package output

import "io"

func detectTerminalWidth(io.Writer) (int, bool) {
	return 0, false
}

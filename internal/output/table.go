package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"golang.org/x/term"
)

const (
	// Cell padding printTable gives tabwriter.
	tablePadding = 2
	// Narrowest a column shrinks to, before its header is taken into account.
	minColumnWidth = 8
	ellipsis       = "..."
)

func (p *Printer) printTable(data Formattable) error {
	headers := data.TableHeaders()
	rows := data.TableRows()

	if len(headers) == 0 {
		return nil
	}

	rows = fitColumns(headers, rows, p.tableWidth())

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, tablePadding, ' ', 0)

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	out := buf.String()
	if ColorEnabled(p.Writer, p.NoColor) {
		// tabwriter sizes a column by the byte count of its cells, so the escape
		// sequences go around the laid out header line instead of each header cell.
		header, body, _ := strings.Cut(out, "\n")
		out = NewColor(true, color.Bold).Sprint(header) + "\n" + body
	}

	_, err := io.WriteString(p.Writer, out)
	return err
}

// tableWidth returns the width the table must fit in. It is 0 when the writer
// has no width of its own, which lets a redirect or a pipe carry every value in
// full.
func (p *Printer) tableWidth() int {
	if p.width > 0 {
		return p.width
	}

	f, ok := p.Writer.(*os.File)
	if !ok {
		return 0
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0
	}
	return width
}

// fitColumns truncates the cells that make the table wider than limit. Only the
// columns before the last one shrink: tabwriter pads every cell of a row except
// the last, so a long value in the last column cannot move the columns before it.
func fitColumns(headers []string, rows [][]string, limit int) [][]string {
	if limit <= 0 || len(headers) < 2 {
		return rows
	}

	widths := make([]int, len(headers))
	floors := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = runeLen(header)
		floors[i] = max(widths[i], minColumnWidth)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				widths[i] = max(widths[i], runeLen(cell))
			}
		}
	}

	total := tablePadding * (len(widths) - 1)
	for _, width := range widths {
		total += width
	}

	for total > limit {
		widest := -1
		for i := range widths[:len(widths)-1] {
			if widths[i] > floors[i] && (widest == -1 || widths[i] > widths[widest]) {
				widest = i
			}
		}
		if widest == -1 {
			break
		}
		widths[widest]--
		total--
	}

	fitted := make([][]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(row))
		copy(cells, row)
		for j := range cells {
			if j < len(widths)-1 {
				cells[j] = truncate(cells[j], widths[j])
			}
		}
		fitted[i] = cells
	}
	return fitted
}

// truncate cuts s to width, giving the last characters to the ellipsis.
func truncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= len(ellipsis) {
		return string(runes[:width])
	}
	return string(runes[:width-len(ellipsis)]) + ellipsis
}

func runeLen(s string) int {
	return len([]rune(s))
}

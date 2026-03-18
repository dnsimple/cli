package output

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	tablePadding          = 2
	maxCompactColumnWidth = 12
)

func (p *Printer) printTable(data Formattable) error {
	headers := data.TableHeaders()
	rows := data.TableRows()

	if len(headers) == 0 {
		return nil
	}

	widths := naturalColumnWidths(headers, rows)
	if tableWidth, ok := p.effectiveTableWidth(); ok {
		widths = fitColumnWidths(widths, headers, tableWidth)
	}

	return writeTable(p.Writer, headers, rows, widths)
}

func (p *Printer) effectiveTableWidth() (int, bool) {
	if p.TableWidth > 0 {
		return p.TableWidth, true
	}

	if width, ok := detectTerminalWidth(p.Writer); ok {
		return width, true
	}

	if raw, ok := os.LookupEnv("COLUMNS"); ok {
		if width, err := strconv.Atoi(raw); err == nil && width > 0 {
			return width, true
		}
	}

	return 0, false
}

func naturalColumnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = cellWidth(header)
	}

	for _, row := range rows {
		for i := 0; i < len(headers) && i < len(row); i++ {
			if width := cellWidth(row[i]); width > widths[i] {
				widths[i] = width
			}
		}
	}

	return widths
}

func fitColumnWidths(widths []int, headers []string, tableWidth int) []int {
	if tableWidth <= 0 || totalTableWidth(widths) <= tableWidth {
		return widths
	}

	fitted := append([]int(nil), widths...)
	minWidths := minimumColumnWidths(widths, headers)

	for totalTableWidth(fitted) > tableWidth {
		column := widestShrinkableColumn(fitted, minWidths)
		if column == -1 {
			break
		}
		fitted[column]--
	}

	return fitted
}

func minimumColumnWidths(widths []int, headers []string) []int {
	minWidths := make([]int, len(widths))
	for i, width := range widths {
		minWidths[i] = maxInt(cellWidth(headers[i]), minInt(width, maxCompactColumnWidth))
	}
	return minWidths
}

func widestShrinkableColumn(widths []int, minWidths []int) int {
	column := -1
	for i := range widths {
		if widths[i] <= minWidths[i] {
			continue
		}
		if column == -1 || widths[i] > widths[column] {
			column = i
		}
	}
	return column
}

func totalTableWidth(widths []int) int {
	if len(widths) == 0 {
		return 0
	}

	total := tablePadding * (len(widths) - 1)
	for _, width := range widths {
		total += width
	}
	return total
}

func writeTable(w io.Writer, headers []string, rows [][]string, widths []int) error {
	if err := writeVisualRow(w, headers, widths); err != nil {
		return err
	}

	for _, row := range rows {
		cells := append([]string(nil), row...)
		if len(cells) < len(headers) {
			cells = append(cells, make([]string, len(headers)-len(cells))...)
		}
		if err := writeVisualRow(w, cells[:len(headers)], widths); err != nil {
			return err
		}
	}

	return nil
}

func writeVisualRow(w io.Writer, row []string, widths []int) error {
	wrapped := make([][]string, len(widths))
	lines := 1

	for i, width := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		wrapped[i] = wrapCell(cell, width)
		if len(wrapped[i]) > lines {
			lines = len(wrapped[i])
		}
	}

	for line := 0; line < lines; line++ {
		var b strings.Builder
		for column, width := range widths {
			if column > 0 {
				b.WriteString(strings.Repeat(" ", tablePadding))
			}

			part := ""
			if line < len(wrapped[column]) {
				part = wrapped[column][line]
			}
			b.WriteString(part)
			if pad := width - utf8.RuneCountInString(part); pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		}

		if _, err := fmt.Fprintln(w, strings.TrimRight(b.String(), " ")); err != nil {
			return err
		}
	}

	return nil
}

func wrapCell(cell string, width int) []string {
	if width <= 0 {
		return []string{cell}
	}

	lines := strings.Split(cell, "\n")
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapLine(line, width)...)
	}

	if len(wrapped) == 0 {
		return []string{""}
	}

	return wrapped
}

func wrapLine(line string, width int) []string {
	if line == "" {
		return []string{""}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	wrapped := make([]string, 0, 1)
	current := ""

	for _, word := range words {
		for utf8.RuneCountInString(word) > width {
			if current != "" {
				wrapped = append(wrapped, current)
				current = ""
			}

			part, rest := splitAtWidth(word, width)
			wrapped = append(wrapped, part)
			word = rest
		}

		if current == "" {
			current = word
			continue
		}

		if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(word) <= width {
			current += " " + word
			continue
		}

		wrapped = append(wrapped, current)
		current = word
	}

	if current != "" {
		wrapped = append(wrapped, current)
	}

	return wrapped
}

func splitAtWidth(s string, width int) (string, string) {
	runes := []rune(s)
	if len(runes) <= width {
		return s, ""
	}
	return string(runes[:width]), string(runes[width:])
}

func cellWidth(cell string) int {
	width := 0
	for _, line := range strings.Split(cell, "\n") {
		if lineWidth := utf8.RuneCountInString(line); lineWidth > width {
			width = lineWidth
		}
	}
	return width
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

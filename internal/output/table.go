package output

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

func (p *Printer) printTable(data Formattable) error {
	headers := data.TableHeaders()
	rows := data.TableRows()

	if len(headers) == 0 {
		return nil
	}

	w := tabwriter.NewWriter(p.Writer, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

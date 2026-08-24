package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

func (p *Printer) printTable(data Formattable) error {
	headers := data.TableHeaders()
	rows := data.TableRows()

	if len(headers) == 0 {
		return nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if !ColorEnabled(p.Writer, p.NoColor) {
		_, err := p.Writer.Write(buf.Bytes())
		return err
	}

	// tabwriter sizes a column by the byte count of its cells, so the escape
	// sequences go around the laid out header line instead of each header cell.
	header, body, _ := strings.Cut(buf.String(), "\n")
	bold := color.New(color.Bold)
	bold.EnableColor()
	if _, err := fmt.Fprintln(p.Writer, bold.Sprint(header)); err != nil {
		return err
	}
	_, err := io.WriteString(p.Writer, body)
	return err
}

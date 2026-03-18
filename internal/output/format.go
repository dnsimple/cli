package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"
)

// Format represents the output format type.
type Format int

const (
	FormatTable    Format = iota // Human-readable table
	FormatJSON                   // Structured JSON
	FormatTemplate               // Go template
)

// Formattable is implemented by types that can be rendered in multiple formats.
type Formattable interface {
	TableHeaders() []string
	TableRows() [][]string
	JSONData() any
}

// Printer handles output rendering.
type Printer struct {
	Writer     io.Writer
	ErrWriter  io.Writer
	Format     Format
	Template   string
	NoColor    bool
	TableWidth int
}

// NewPrinter creates a new Printer with the given format settings.
func NewPrinter(writer io.Writer, errWriter io.Writer, format Format, templateStr string, noColor bool) *Printer {
	return &Printer{
		Writer:    writer,
		ErrWriter: errWriter,
		Format:    format,
		Template:  templateStr,
		NoColor:   noColor,
	}
}

// Print renders the given data in the configured format.
func (p *Printer) Print(data Formattable) error {
	switch p.Format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatTemplate:
		return p.printTemplate(data)
	default:
		return p.printTable(data)
	}
}

func (p *Printer) printJSON(data Formattable) error {
	enc := json.NewEncoder(p.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data.JSONData())
}

func (p *Printer) printTemplate(data Formattable) error {
	tmpl, err := template.New("output").Parse(p.Template)
	if err != nil {
		return fmt.Errorf("invalid format template: %w", err)
	}
	return tmpl.Execute(p.Writer, data.JSONData())
}

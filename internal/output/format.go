package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/fatih/color"
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
	TemplateData() any
}

// PageInfo describes pagination state for the table discovery hint. It is
// intentionally SDK-free so the output package stays decoupled from the API client.
type PageInfo struct {
	Noun         string // resource label, e.g. "records"
	Shown        int    // items on the current page
	CurrentPage  int
	TotalPages   int
	TotalEntries int
	CanFetchAll  bool // command exposes an --all flag
	CanPaginate  bool // command exposes --page/--per-page flags
}

// Printer handles output rendering.
type Printer struct {
	Writer    io.Writer
	ErrWriter io.Writer
	Format    Format
	Template  string
	NoColor   bool
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

// PrintList renders a paginated list. For table output spanning more than one
// page it writes a discovery hint to the error stream above the table, so a user
// (or AI agent) can see that more results exist and how to reach them. JSON and
// template output are left untouched: JSON already embeds the pagination object.
func (p *Printer) PrintList(data Formattable, info *PageInfo) error {
	hint := p.Format == FormatTable && info != nil && info.TotalPages > 1 && p.ErrWriter != nil

	// The hints are advice about the table, so they are faint and the table is not.
	faint := color.New(color.Faint)
	faint.DisableColor()
	if hint && ColorEnabled(p.ErrWriter, p.NoColor) {
		faint.EnableColor()
	}

	if hint {
		// Summary above the table so it stays visible before a long list scrolls past.
		fmt.Fprintf(p.ErrWriter, "%s\n\n", faint.Sprintf("Showing %d of %d %s (page %d of %d).",
			info.Shown, info.TotalEntries, info.Noun, info.CurrentPage, info.TotalPages))
	}
	if err := p.Print(data); err != nil {
		return err
	}
	if hint {
		// Navigation advice below the table, where it lands next to the prompt.
		if nav := info.navHint(); nav != "" {
			fmt.Fprintf(p.ErrWriter, "\n%s\n", faint.Sprint(nav))
		}
	}
	return nil
}

// navHint returns advice on retrieving more results, naming only the flags the
// command actually defines. It is empty when the command exposes no way to page.
func (info *PageInfo) navHint() string {
	switch {
	case info.CanFetchAll:
		return "Pass --all to fetch every page, or --page <n>/--per-page <n> to navigate."
	case info.CanPaginate:
		return "Pass --page <n> or --per-page <n> to see more."
	default:
		return ""
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
	return tmpl.Execute(p.Writer, data.TemplateData())
}

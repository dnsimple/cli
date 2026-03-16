package output

import (
	"bytes"
	"strings"
	"testing"
)

type stubFormattable struct {
	headers []string
	rows    [][]string
	json    any
}

func (s *stubFormattable) TableHeaders() []string { return s.headers }
func (s *stubFormattable) TableRows() [][]string  { return s.rows }
func (s *stubFormattable) JSONData() any          { return s.json }

func TestPrinterPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatJSON}

	err := p.Print(&stubFormattable{
		headers: []string{"NAME"},
		rows:    [][]string{{"dnsimple"}},
		json:    map[string]string{"name": "dnsimple"},
	})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "{\n  \"name\": \"dnsimple\"\n}\n"
	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestPrinterPrintTemplate(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTemplate, Template: "{{.name}}/{{.id}}"}

	err := p.Print(&stubFormattable{json: map[string]any{"name": "dnsimple", "id": 1010}})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if got := buf.String(); got != "dnsimple/1010" {
		t.Fatalf("Print() = %q, want %q", got, "dnsimple/1010")
	}
}

func TestPrinterPrintTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTable}

	err := p.Print(&stubFormattable{
		headers: []string{"NAME", "VALUE"},
		rows: [][]string{
			{"alpha", "1"},
			{"beta", "22"},
		},
	})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	out := buf.String()
	for _, part := range []string{"NAME", "VALUE", "alpha", "beta", "22"} {
		if !strings.Contains(out, part) {
			t.Fatalf("table output %q does not contain %q", out, part)
		}
	}
}

func TestPrinterPrintTableEmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTable}

	err := p.Print(&stubFormattable{})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("buffer length = %d, want 0", buf.Len())
	}
}

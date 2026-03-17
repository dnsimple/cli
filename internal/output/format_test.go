package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	if !assert.NoError(t, err) {
		return
	}

	want := "{\n  \"name\": \"dnsimple\"\n}\n"
	assert.Equal(t, want, buf.String())
}

func TestPrinterPrintTemplate(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTemplate, Template: "{{.name}}/{{.id}}"}

	err := p.Print(&stubFormattable{json: map[string]any{"name": "dnsimple", "id": 1010}})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "dnsimple/1010", buf.String())
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
	if !assert.NoError(t, err) {
		return
	}

	out := buf.String()
	for _, part := range []string{"NAME", "VALUE", "alpha", "beta", "22"} {
		assert.True(t, strings.Contains(out, part), "table output %q does not contain %q", out, part)
	}
}

func TestPrinterPrintTableEmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTable}

	err := p.Print(&stubFormattable{})
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, buf.Len())
}

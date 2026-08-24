package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubFormattable struct {
	headers  []string
	rows     [][]string
	json     any
	template any
}

func (s *stubFormattable) TableHeaders() []string { return s.headers }
func (s *stubFormattable) TableRows() [][]string  { return s.rows }
func (s *stubFormattable) JSONData() any          { return s.json }
func (s *stubFormattable) TemplateData() any {
	if s.template != nil {
		return s.template
	}
	return s.json
}

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

	err := p.Print(&stubFormattable{
		json:     map[string]any{"data": map[string]any{"name": "wrapped", "id": 0}},
		template: map[string]any{"name": "dnsimple", "id": 1010},
	})
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

func TestPrinterPrintListHintsStayPlainWithoutATerminal(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	err := p.PrintList(listData(), &PageInfo{
		Noun: "records", Shown: 2, CurrentPage: 1, TotalPages: 3, TotalEntries: 6, CanFetchAll: true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.NotContains(t, errOut.String(), "\x1b[")
	assert.Contains(t, errOut.String(), "Showing 2 of 6 records (page 1 of 3).")
	assert.Contains(t, errOut.String(), "Pass --all to fetch every page")
}

func TestPrinterPrintTableStaysPlainWithoutATerminal(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Writer: &buf, Format: FormatTable}

	err := p.Print(&stubFormattable{
		headers: []string{"ID", "NAME"},
		rows:    [][]string{{"1", "example.com"}},
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "ID  NAME\n1   example.com\n", buf.String())
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

func listData() *stubFormattable {
	return &stubFormattable{
		headers: []string{"NAME"},
		rows:    [][]string{{"alpha"}, {"beta"}},
		json:    map[string]string{"name": "alpha"},
	}
}

func TestPrinterPrintListShowsHintOnTable(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	err := p.PrintList(listData(), &PageInfo{
		Noun: "records", Shown: 30, CurrentPage: 1, TotalPages: 5, TotalEntries: 142, CanFetchAll: true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, errOut.String(), "Showing 30 of 142 records (page 1 of 5)")
	assert.Contains(t, errOut.String(), "--all")
	assert.Contains(t, out.String(), "alpha")
}

func TestPrinterPrintListHintOmitsAllFlagWhenUnsupported(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	err := p.PrintList(listData(), &PageInfo{
		Noun: "certificates", Shown: 30, CurrentPage: 1, TotalPages: 5, TotalEntries: 142, CanFetchAll: false, CanPaginate: true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, errOut.String(), "Showing 30 of 142 certificates (page 1 of 5)")
	assert.NotContains(t, errOut.String(), "--all")
	assert.Contains(t, errOut.String(), "--page")
}

func TestPrinterPrintListNoNavAdviceWhenNoPaginationFlags(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	// Command exposes neither --all nor --page: show the summary but never
	// advise a flag that does not exist on the command.
	err := p.PrintList(listData(), &PageInfo{
		Noun: "services", Shown: 30, CurrentPage: 1, TotalPages: 2, TotalEntries: 41,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, errOut.String(), "Showing 30 of 41 services (page 1 of 2)")
	assert.NotContains(t, errOut.String(), "--page")
	assert.NotContains(t, errOut.String(), "--all")
}

func TestPrinterPrintListNoHintOnSinglePage(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	err := p.PrintList(listData(), &PageInfo{
		Noun: "records", Shown: 2, CurrentPage: 1, TotalPages: 1, TotalEntries: 2,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, errOut.Len())
	assert.Contains(t, out.String(), "alpha")
}

func TestPrinterPrintListNoHintWhenInfoNil(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatTable}

	err := p.PrintList(listData(), nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, errOut.Len())
}

func TestPrinterPrintListNoHintForJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	p := &Printer{Writer: &out, ErrWriter: &errOut, Format: FormatJSON}

	err := p.PrintList(listData(), &PageInfo{
		Noun: "records", Shown: 30, CurrentPage: 1, TotalPages: 5, TotalEntries: 142, CanFetchAll: true,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, errOut.Len())
	assert.Contains(t, out.String(), "alpha")
}

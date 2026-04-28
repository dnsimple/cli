package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeGroupings(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantOrdered    []string
		wantCanonical  string
		wantErrSubstrs []string
	}{
		{
			name:          "empty input",
			input:         "",
			wantOrdered:   nil,
			wantCanonical: "",
		},
		{
			name:          "whitespace only",
			input:         "   ",
			wantOrdered:   nil,
			wantCanonical: "",
		},
		{
			name:          "single zone_name",
			input:         "zone_name",
			wantOrdered:   []string{"zone_name"},
			wantCanonical: "zone_name",
		},
		{
			name:          "single date",
			input:         "date",
			wantOrdered:   []string{"date"},
			wantCanonical: "date",
		},
		{
			name:          "zone_name then date preserves order",
			input:         "zone_name,date",
			wantOrdered:   []string{"zone_name", "date"},
			wantCanonical: "zone_name,date",
		},
		{
			name:          "date then zone_name preserves order",
			input:         "date,zone_name",
			wantOrdered:   []string{"date", "zone_name"},
			wantCanonical: "date,zone_name",
		},
		{
			name:          "duplicate keys are deduped",
			input:         "zone_name,zone_name",
			wantOrdered:   []string{"zone_name"},
			wantCanonical: "zone_name",
		},
		{
			name:          "whitespace around keys is tolerated",
			input:         " zone_name , date ",
			wantOrdered:   []string{"zone_name", "date"},
			wantCanonical: "zone_name,date",
		},
		{
			name:           "unknown grouping is rejected",
			input:          "foo",
			wantErrSubstrs: []string{`"foo"`, "zone_name", "date"},
		},
		{
			name:           "alias zone is rejected",
			input:          "zone",
			wantErrSubstrs: []string{`"zone"`, "zone_name", "date"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ordered, canonical, err := normalizeGroupings(tt.input)
			if len(tt.wantErrSubstrs) > 0 {
				if assert.Error(t, err) {
					for _, substr := range tt.wantErrSubstrs {
						assert.Contains(t, err.Error(), substr)
					}
				}
				assert.Nil(t, ordered)
				assert.Equal(t, "", canonical)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOrdered, ordered)
			assert.Equal(t, tt.wantCanonical, canonical)
		})
	}
}

func TestAnalyticsQueryNoGroupings(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/dns_analytics", r.URL.Path)
		assert.Equal(t, "", r.URL.Query().Get("groupings"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"headers":["volume"],"rows":[[12000000000]]},"query":{"account_id":1950,"start_date":"","end_date":"","sort":"","page":0,"per_page":100,"groupings":""},"pagination":{"current_page":1,"per_page":100,"total_entries":1,"total_pages":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newAnalyticsQueryCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	out := stdout.String()
	assert.Contains(t, out, "VOLUME")
	assert.NotContains(t, out, "ZONE")
	assert.NotContains(t, out, "DATE")
	assert.Contains(t, out, "12000000000")
}

func TestAnalyticsQuerySingleGrouping(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/dns_analytics", r.URL.Path)
		assert.Equal(t, "zone_name", r.URL.Query().Get("groupings"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"headers":["zone_name","volume"],"rows":[["example.com",1200],["foo.com",3400]]},"query":{"account_id":1950,"start_date":"","end_date":"","sort":"","page":0,"per_page":100,"groupings":"zone_name"},"pagination":{"current_page":1,"per_page":100,"total_entries":2,"total_pages":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newAnalyticsQueryCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--groupings", "zone_name"})

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	out := stdout.String()
	assert.Contains(t, out, "ZONE")
	assert.Contains(t, out, "VOLUME")
	assert.NotContains(t, out, "DATE")
	assert.Contains(t, out, "example.com")
	assert.Contains(t, out, "1200")
	assert.Contains(t, out, "foo.com")
	assert.Contains(t, out, "3400")
}

func TestAnalyticsQueryMultipleGroupings(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/dns_analytics", r.URL.Path)
		assert.Equal(t, "zone_name,date", r.URL.Query().Get("groupings"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"headers":["zone_name","date","volume"],"rows":[["example.com","2026-03-15",1200],["example.com","2026-03-16",1300]]},"query":{"account_id":1950,"start_date":"","end_date":"","sort":"","page":0,"per_page":100,"groupings":"zone_name,date"},"pagination":{"current_page":1,"per_page":100,"total_entries":2,"total_pages":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newAnalyticsQueryCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--groupings", "zone_name,date"})

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	out := stdout.String()
	zoneIdx := bytes.Index(stdout.Bytes(), []byte("ZONE"))
	dateIdx := bytes.Index(stdout.Bytes(), []byte("DATE"))
	volumeIdx := bytes.Index(stdout.Bytes(), []byte("VOLUME"))
	assert.True(t, zoneIdx >= 0 && dateIdx > zoneIdx && volumeIdx > dateIdx, "expected headers in order ZONE, DATE, VOLUME, got: %q", out)
	assert.Contains(t, out, "example.com")
	assert.Contains(t, out, "2026-03-15")
	assert.Contains(t, out, "1200")
	assert.Contains(t, out, "2026-03-16")
	assert.Contains(t, out, "1300")
}

func TestAnalyticsQueryInvalidGroupingDoesNotCallAPI(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to %s", r.URL.Path)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newAnalyticsQueryCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--groupings", "foo"})

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `"foo"`)
		assert.Contains(t, err.Error(), "zone_name")
		assert.Contains(t, err.Error(), "date")
	}
	assert.Zero(t, stdout.Len())
}

package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestResearchStatus_Available(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/domains/research/status", r.URL.Path)
		assert.Equal(t, "available.com", r.URL.Query().Get("domain"))
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"request_id":"abc-123","domain":"available.com","availability":"available","errors":[]}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newResearchStatusCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"available.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "available.com")
	assert.Contains(t, stdout.String(), "available")
}

func TestResearchStatus_Unavailable(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/domains/research/status", r.URL.Path)
		assert.Equal(t, "taken.com", r.URL.Query().Get("domain"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"request_id":"def-456","domain":"taken.com","availability":"unavailable","errors":[]}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newResearchStatusCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"taken.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "taken.com")
	assert.Contains(t, stdout.String(), "unavailable")
}

func TestResearchStatus_UnsupportedTLD(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"request_id":"ghi-789","domain":"taken.co.fa","availability":"unknown","errors":["TLD not supported for registration"]}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newResearchStatusCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"taken.co.fa"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "unknown")
	assert.Contains(t, stdout.String(), "TLD not supported for registration")
}

func TestResearchStatus_JSON(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"request_id":"abc-123","domain":"example.com","availability":"available","errors":[]}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }
	f.Flags.JSON = true

	cmd := newResearchStatusCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"example.com"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), `"availability": "available"`)
	assert.Contains(t, stdout.String(), `"request_id": "abc-123"`)
}

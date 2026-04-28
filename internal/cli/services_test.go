package cli

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestServicesApplyUsesDomainThenService(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/domains/example.com/services/github-pages", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Empty(t, r.URL.RawQuery)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"Settings":{"branch":"main","repo":"dnsimple/site"}}`, string(body))

		w.WriteHeader(http.StatusNoContent)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newServicesApplyCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"example.com", "github-pages", "--settings", "repo=dnsimple/site", "--settings", "branch=main"})

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Service github-pages applied to example.com\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestServicesUnapplyUsesDomainThenService(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/domains/example.com/services/github-pages", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newServicesUnapplyCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"example.com", "github-pages", "--yes"})

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Service github-pages removed from example.com\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestServicesApplyAndUnapplyUseStringsAreDomainFirst(t *testing.T) {
	f := cmdutil.NewFactory("test")

	assert.Equal(t, "apply <domain> <service>", newServicesApplyCmd(f).Use)
	assert.Equal(t, "unapply <domain> <service>", newServicesUnapplyCmd(f).Use)
}

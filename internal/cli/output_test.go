package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	internalclient "github.com/dnsimple/dnsimple-cli/internal/client"
	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestWhoamiUsesTemplateOutputOnCommandStream(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/whoami", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"user":{"id":101,"email":"user@example.com"},"account":{"id":202,"email":"account@example.com"}}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.Flags.Format = "{{.UserEmail}}/{{.AccountEmail}}"

	cmd := newWhoamiCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "user@example.com/account@example.com", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestAuthStatusUsesTemplateOutputOnCommandStream(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/whoami":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":{"user":{"id":101,"email":"user@example.com"},"account":{"id":202,"email":"account@example.com"}}}`)
		case "/v2/accounts":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[{"id":202,"email":"account@example.com"}]}`)
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
		}
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "202", nil }
	f.Flags.Format = "{{.Environment}}/{{.UserEmail}}/{{.AccountEmail}}"

	cmd := newAuthStatusCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "production/user@example.com/account@example.com", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestAuthStatusWarnsWhenDefaultAccountIsNotAccessible(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/whoami":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":{"user":{"id":101,"email":"user@example.com"},"account":null}}`)
		case "/v2/accounts":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[{"id":202,"email":"account@example.com"}]}`)
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
		}
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "999", nil }
	f.Flags.Format = "{{.AccountID}}/{{.AccountEmail}}/{{.Warning}}"

	cmd := newAuthStatusCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "999//account 999 is not accessible with the current token; run 'dnsimple auth switch <account-id>' to update", stdout.String())
}

func testCLIClient(t *testing.T, handler http.HandlerFunc) (*dnsimple.Client, *config.Config) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	cfg := &config.Config{BaseURL: server.URL}
	return internalclient.NewClient(cfg, "token", "test"), cfg
}

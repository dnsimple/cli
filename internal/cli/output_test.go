package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	internalclient "github.com/dnsimple/cli/internal/client"
	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
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
	f.Context = func() (*config.ResolvedContext, error) {
		return &config.ResolvedContext{
			ContextName: "personal",
			BaseURL:     cfg.BaseURL,
			Host:        config.ProductionHost,
			Token:       "tok",
			AccountID:   "202",
		}, nil
	}
	f.Flags.Format = "{{.Context}}/{{.Environment}}/{{.UserEmail}}/{{.AccountEmail}}"

	cmd := newAuthStatusCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "personal/production/user@example.com/account@example.com", stdout.String())
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
	f.Context = func() (*config.ResolvedContext, error) {
		return &config.ResolvedContext{
			ContextName: "stale",
			BaseURL:     cfg.BaseURL,
			Host:        config.ProductionHost,
			Token:       "tok",
			AccountID:   "999",
		}, nil
	}
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

	assert.Equal(t, "999//account 999 is not accessible with the current token; run 'dnsimple auth login' to refresh the context", stdout.String())
}

func TestDomainsGetUsesTemplateOutputOnUnderlyingResourceFields(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/domains/example.com", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":101,"name":"example.com","state":"registered","auto_renew":true,"private_whois":false,"expires_at":"2026-01-01","created_at":"2025-01-01","updated_at":"2025-06-01"}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }
	f.Flags.Format = "{{.Name}}/{{.ID}}"

	cmd := newDomainsGetCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"example.com"})

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "example.com/101", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestDomainsListUsesTemplateOutputOnUnderlyingResourceList(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/domains", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":101,"name":"alpha.example","state":"registered","auto_renew":true,"private_whois":false,"expires_at":"2026-01-01","created_at":"2025-01-01","updated_at":"2025-06-01"},{"id":102,"name":"beta.example","state":"registered","auto_renew":false,"private_whois":false,"expires_at":"2026-02-01","created_at":"2025-02-01","updated_at":"2025-06-02"}],"pagination":{"current_page":1,"per_page":30,"total_entries":2,"total_pages":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }
	f.Flags.Format = `{{range .}}{{.Name}}{{"\n"}}{{end}}`

	cmd := newDomainsListCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "alpha.example\nbeta.example\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestRecordsListShowsPaginationHintOnTable(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1950/zones/example.com/records", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":1,"type":"A","name":"","content":"1.2.3.4","ttl":3600},{"id":2,"type":"A","name":"www","content":"1.2.3.5","ttl":3600}],"pagination":{"current_page":1,"per_page":30,"total_entries":142,"total_pages":5}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newRecordsListCmd(f)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, []string{"example.com"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, stderr.String(), "Showing 2 of 142 records (page 1 of 5)")
	assert.Contains(t, stderr.String(), "--all")
	assert.Contains(t, stdout.String(), "1.2.3.4")
}

func TestRecordsListJSONIncludesPaginationWithoutHint(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":1,"type":"A","name":"","content":"1.2.3.4","ttl":3600}],"pagination":{"current_page":1,"per_page":30,"total_entries":142,"total_pages":5}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }
	f.Flags.JSON = true

	cmd := newRecordsListCmd(f)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, []string{"example.com"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, stderr.Len())
	assert.Contains(t, stdout.String(), `"pagination"`)
	assert.Contains(t, stdout.String(), `"total_entries": 142`)
}

func TestRecordsListNoHintOnSinglePage(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":1,"type":"A","name":"","content":"1.2.3.4","ttl":3600}],"pagination":{"current_page":1,"per_page":30,"total_entries":1,"total_pages":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newRecordsListCmd(f)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, []string{"example.com"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Zero(t, stderr.Len())
	assert.Contains(t, stdout.String(), "1.2.3.4")
}

// A paginated list command must expose the navigation flags its hint advertises,
// otherwise the hint points users at flags that error with "unknown flag".
func TestPaginatedListCommandsExposeNavigationFlags(t *testing.T) {
	f := cmdutil.NewFactory("test")
	commands := map[string]*cobra.Command{
		"services list":          newServicesListCmd(f),
		"services applied":       newServicesAppliedCmd(f),
		"templates list":         newTemplatesListCmd(f),
		"template records list":  newTemplateRecordsListCmd(f),
		"email-forwards list":    newEmailForwardsListCmd(f),
		"ds-records list":        newDsRecordsListCmd(f),
		"pushes list":            newPushesListCmd(f),
		"registrant-change list": newRegistrantChangeListCmd(f),
		"records list":           newRecordsListCmd(f),
		"domains list":           newDomainsListCmd(f),
		"zones list":             newZonesListCmd(f),
		"contacts list":          newContactsListCmd(f),
		"certificates list":      newCertsListCmd(f),
		"tlds list":              newTldsListCmd(f),
		"analytics query":        newAnalyticsQueryCmd(f),
	}
	for name, cmd := range commands {
		for _, flag := range []string{"all", "page", "per-page"} {
			assert.NotNilf(t, cmd.Flags().Lookup(flag), "%s should define --%s", name, flag)
		}
	}
}

func testCLIClient(t *testing.T, handler http.HandlerFunc) (*dnsimple.Client, *config.Config) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	cfg := &config.Config{BaseURL: server.URL}
	c := internalclient.New(internalclient.Options{
		BaseURL: server.URL,
		Token:   "token",
		Version: "test",
	})
	return c, cfg
}

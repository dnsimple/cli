package cli

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestConfirmRegisteredDomainDeletionInputAcceptsMatchingDomain(t *testing.T) {
	var stderr bytes.Buffer

	err := confirmRegisteredDomainDeletionInput(strings.NewReader("example.com\n"), &stderr, true, "example.com", false, false)
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, stderr.String(), "WARNING: example.com is currently registered.")
	assert.Contains(t, stderr.String(), "downgrade the domain to hosted")
	assert.Contains(t, stderr.String(), "Type the domain name to continue:")
}

func TestConfirmRegisteredDomainDeletionInputRejectsMismatchedDomain(t *testing.T) {
	var stderr bytes.Buffer

	err := confirmRegisteredDomainDeletionInput(strings.NewReader("wrong.example\n"), &stderr, true, "example.com", false, false)
	assert.EqualError(t, err, "confirmation declined")
	assert.Contains(t, stderr.String(), "Type the domain name to continue:")
}

func TestDomainsDeleteRegisteredDomainRequiresExtraAcknowledgement(t *testing.T) {
	requests := make([]string, 0, 2)
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		assert.Equal(t, "/v2/1950/domains/example.com", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"id":101,"name":"example.com","state":"registered","auto_renew":true,"private_whois":false,"expires_at":"2026-01-01","created_at":"2025-01-01","updated_at":"2025-06-01"}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newDomainsDeleteCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}

	err := cmd.RunE(cmd, []string{"example.com"})
	assert.EqualError(t, err, "domain example.com is registered; rerun with --yes --confirm-registered-domain to acknowledge the downgrade to hosted and loss of registration metadata")
	assert.Equal(t, []string{"GET /v2/1950/domains/example.com"}, requests)
	assert.Zero(t, stdout.Len())
	assert.Zero(t, stderr.Len())
}

func TestDomainsDeleteRegisteredDomainWithYesAndConfirmFlagDeletes(t *testing.T) {
	requests := make([]string, 0, 2)
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		assert.Equal(t, "/v2/1950/domains/example.com", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"id":101,"name":"example.com","state":"registered","auto_renew":true,"private_whois":false,"expires_at":"2026-01-01","created_at":"2025-01-01","updated_at":"2025-06-01"}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newDomainsDeleteCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}
	if err := cmd.Flags().Set("confirm-registered-domain", "true"); err != nil {
		t.Fatalf("set confirm-registered-domain: %v", err)
	}

	err := cmd.RunE(cmd, []string{"example.com"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{
		"GET /v2/1950/domains/example.com",
		"DELETE /v2/1950/domains/example.com",
	}, requests)
	assert.Equal(t, "Domain example.com deleted\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestDomainsDeleteHostedDomainOnlyRequiresYes(t *testing.T) {
	requests := make([]string, 0, 2)
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		assert.Equal(t, "/v2/1950/domains/example.com", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"id":101,"name":"example.com","state":"hosted","auto_renew":false,"private_whois":false,"expires_at":"","created_at":"2025-01-01","updated_at":"2025-06-01"}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newDomainsDeleteCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}

	err := cmd.RunE(cmd, []string{"example.com"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{
		"GET /v2/1950/domains/example.com",
		"DELETE /v2/1950/domains/example.com",
	}, requests)
	assert.Equal(t, "Domain example.com deleted\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

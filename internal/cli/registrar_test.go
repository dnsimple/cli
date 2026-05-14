package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestParseExtendedAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]string
	}{
		{
			name:  "nil input returns nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty input returns nil",
			input: []string{},
			want:  nil,
		},
		{
			name:  "single key=value",
			input: []string{"x-accept-ssl-requirement=1"},
			want:  map[string]string{"x-accept-ssl-requirement": "1"},
		},
		{
			name:  "multiple key=value pairs",
			input: []string{"key1=val1", "key2=val2", "key3=val3"},
			want:  map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"},
		},
		{
			name:  "value containing equals sign",
			input: []string{"key=val=ue"},
			want:  map[string]string{"key": "val=ue"},
		},
		{
			name:  "empty value",
			input: []string{"key="},
			want:  map[string]string{"key": ""},
		},
		{
			name:  "missing separator is skipped",
			input: []string{"noequals"},
			want:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExtendedAttributes(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistrarRegister(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/registrar/domains/example.com/registrations", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, _ := io.ReadAll(r.Body)
		var input map[string]any
		_ = json.Unmarshal(body, &input)

		assert.Equal(t, float64(5678), input["registrant_id"])
		assert.Nil(t, input["extended_attributes"])

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":2,"domain_id":200,"registrant_id":5678,"state":"new","auto_renew":true,"whois_privacy":false,"period":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newRegistrarRegisterCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"example.com", "--registrant-id", "5678"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "5678")
}

func TestRegistrarRegisterWithExtendedAttributes(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/registrar/domains/example.app/registrations", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, _ := io.ReadAll(r.Body)
		var input map[string]any
		_ = json.Unmarshal(body, &input)

		assert.Equal(t, float64(1234), input["registrant_id"])
		assert.Equal(t, map[string]any{"x-accept-ssl-requirement": "1"}, input["extended_attributes"])

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":1,"domain_id":100,"registrant_id":1234,"state":"new","auto_renew":true,"whois_privacy":false,"period":1}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newRegistrarRegisterCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"example.app", "--registrant-id", "1234", "--extended-attribute", "x-accept-ssl-requirement=1"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "1234")
}

func TestRegistrarTransfer(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/registrar/domains/example.com/transfers", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, _ := io.ReadAll(r.Body)
		var input map[string]any
		_ = json.Unmarshal(body, &input)

		assert.Equal(t, float64(5678), input["registrant_id"])
		assert.Nil(t, input["extended_attributes"])

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":1,"domain_id":200,"registrant_id":5678,"state":"transferring","auto_renew":true,"whois_privacy":false}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newRegistrarTransferCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"example.com", "--registrant-id", "5678"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "5678")
}

func TestRegistrarTransferWithExtendedAttributes(t *testing.T) {
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/1/registrar/domains/example.us/transfers", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, _ := io.ReadAll(r.Body)
		var input map[string]any
		_ = json.Unmarshal(body, &input)

		assert.Equal(t, float64(1234), input["registrant_id"])
		assert.Equal(t, map[string]any{"us_nexus": "C11", "us_purpose": "P3"}, input["extended_attributes"])

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":2,"domain_id":300,"registrant_id":1234,"state":"transferring","auto_renew":true,"whois_privacy":false}}`)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1", nil }

	cmd := newRegistrarTransferCmd(f)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"example.us", "--registrant-id", "1234", "--extended-attribute", "us_nexus=C11", "--extended-attribute", "us_purpose=P3"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "1234")
}

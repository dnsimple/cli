package cli

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dnsimple/dnsimple-cli/internal/cmdutil"
	"github.com/dnsimple/dnsimple-cli/internal/config"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfirmDestructiveActionInputAcceptsPromptedYes(t *testing.T) {
	var stderr bytes.Buffer

	err := confirmDestructiveActionInput(strings.NewReader("yes\n"), &stderr, true, false, "Delete record 123 from example.com?")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Delete record 123 from example.com? [y/N]: ", stderr.String())
}

func TestConfirmDestructiveActionInputRejectsPromptedNo(t *testing.T) {
	var stderr bytes.Buffer

	err := confirmDestructiveActionInput(strings.NewReader("n\n"), &stderr, true, false, "Delete record 123 from example.com?")
	assert.EqualError(t, err, "confirmation declined")
	assert.Equal(t, "Delete record 123 from example.com? [y/N]: ", stderr.String())
}

func TestConfirmDestructiveActionInputRejectsNonInteractiveWithoutYes(t *testing.T) {
	err := confirmDestructiveActionInput(strings.NewReader(""), io.Discard, false, false, "Delete record 123 from example.com?")
	assert.EqualError(t, err, "destructive operation requires confirmation; rerun with --yes")
}

func TestRecordsDeleteRequiresYesWhenNonInteractive(t *testing.T) {
	requests := 0
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNoContent)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newRecordsDeleteCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.RunE(cmd, []string{"example.com", "123"})
	assert.EqualError(t, err, "destructive operation requires confirmation; rerun with --yes")
	assert.Zero(t, requests)
	assert.Zero(t, stdout.Len())
	assert.Zero(t, stderr.Len())
}

func TestRecordsDeleteWithYesDeletesWithoutPrompt(t *testing.T) {
	requests := 0
	client, cfg := testCLIClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	f := cmdutil.NewFactory("test")
	f.Client = func() (*dnsimple.Client, error) { return client, nil }
	f.Config = func() (*config.Config, error) { return cfg, nil }
	f.AccountID = func() (string, error) { return "1950", nil }

	cmd := newRecordsDeleteCmd(f)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}

	err := cmd.RunE(cmd, []string{"example.com", "123"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 1, requests)
	assert.Equal(t, "Record 123 deleted from zone example.com\n", stdout.String())
	assert.Zero(t, stderr.Len())
}

func TestDestructiveResourceCommandsExposeYesFlag(t *testing.T) {
	f := cmdutil.NewFactory("test")

	tests := []struct {
		name  string
		build func(*cmdutil.Factory) *cobra.Command
	}{
		{name: "domains delete", build: newDomainsDeleteCmd},
		{name: "records delete", build: newRecordsDeleteCmd},
		{name: "domains ds-records delete", build: newDsRecordsDeleteCmd},
		{name: "domains email-forwards delete", build: newEmailForwardsDeleteCmd},
		{name: "contacts delete", build: newContactsDeleteCmd},
		{name: "templates delete", build: newTemplatesDeleteCmd},
		{name: "templates records delete", build: newTemplateRecordsDeleteCmd},
		{name: "webhooks delete", build: newWebhooksDeleteCmd},
		{name: "services unapply", build: newServicesUnapplyCmd},
		{name: "registrar registrant-change delete", build: newRegistrantChangeDeleteCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.build(f).Flags().Lookup("yes")
			if !assert.NotNil(t, flag) {
				return
			}
			assert.Equal(t, "y", flag.Shorthand)
			assert.Equal(t, destructiveYesFlagHelp, flag.Usage)
		})
	}
}

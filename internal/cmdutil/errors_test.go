package cmdutil

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
)

func apiError(status int, message string, attrs map[string][]string) error {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	return &dnsimple.ErrorResponse{
		Response: dnsimple.Response{
			HTTPResponse: &http.Response{
				StatusCode: status,
				Request:    req,
			},
		},
		Message:         message,
		AttributeErrors: attrs,
	}
}

func TestFormatAPIErrorWithGenericError(t *testing.T) {
	var buf bytes.Buffer

	FormatAPIError(&buf, errors.New("plain failure"))

	if got := buf.String(); got != "Error: plain failure\n" {
		t.Fatalf("FormatAPIError() = %q, want %q", got, "Error: plain failure\n")
	}
}

func TestFormatAPIErrorUnauthorized(t *testing.T) {
	var buf bytes.Buffer

	FormatAPIError(&buf, apiError(http.StatusUnauthorized, "unauthorized", nil))

	out := buf.String()
	if out != "Error: Authentication failed. Your token may be invalid or expired.\n\nRun 'dnsimple auth login' to re-authenticate.\n" {
		t.Fatalf("FormatAPIError() = %q", out)
	}
}

func TestFormatAPIErrorValidationErrors(t *testing.T) {
	var buf bytes.Buffer

	FormatAPIError(&buf, apiError(http.StatusUnprocessableEntity, "Validation failed", map[string][]string{
		"name": {"can't be blank", "is invalid"},
	}))

	out := buf.String()
	for _, part := range []string{"Error: Validation failed", "name: can't be blank, is invalid"} {
		if !bytes.Contains(buf.Bytes(), []byte(part)) {
			t.Fatalf("output %q does not contain %q", out, part)
		}
	}
}

package cmdutil

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, "Error: plain failure\n", buf.String())
}

func TestFormatAPIErrorUnauthorized(t *testing.T) {
	var buf bytes.Buffer

	FormatAPIError(&buf, apiError(http.StatusUnauthorized, "unauthorized", nil))

	out := buf.String()
	assert.Equal(t, "Error: Authentication failed. Your token may be invalid or expired.\n\nRun 'dnsimple auth login' to re-authenticate.\n", out)
}

func TestFormatAPIErrorValidationErrors(t *testing.T) {
	var buf bytes.Buffer

	FormatAPIError(&buf, apiError(http.StatusUnprocessableEntity, "Validation failed", map[string][]string{
		"name": {"can't be blank", "is invalid"},
	}))

	out := buf.String()
	for _, part := range []string{"Error: Validation failed", "name: can't be blank, is invalid"} {
		assert.Contains(t, out, part)
	}
}

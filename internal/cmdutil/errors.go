package cmdutil

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
)

const (
	ExitOK        = 0
	ExitError     = 1
	ExitUsage     = 2
	ExitAuth      = 4
)

// FormatAPIError formats a DNSimple API error for human-readable display.
func FormatAPIError(w io.Writer, err error) {
	var apiErr *dnsimple.ErrorResponse
	if !errors.As(err, &apiErr) {
		fmt.Fprintf(w, "Error: %s\n", err)
		return
	}

	resp := apiErr.HTTPResponse
	switch {
	case resp != nil && resp.StatusCode == 401:
		fmt.Fprintln(w, "Error: Authentication failed. Your token may be invalid or expired.")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Run 'dnsimple auth login' to re-authenticate.")
	case resp != nil && resp.StatusCode == 404:
		fmt.Fprintf(w, "Error: %s\n", apiErr.Message)
	case resp != nil && resp.StatusCode == 429:
		fmt.Fprintln(w, "Error: API rate limit exceeded.")
	case resp != nil && (resp.StatusCode == 400 || resp.StatusCode == 422):
		fmt.Fprintf(w, "Error: %s\n", apiErr.Message)
		if len(apiErr.AttributeErrors) > 0 {
			fmt.Fprintln(w, "")
			for field, messages := range apiErr.AttributeErrors {
				fmt.Fprintf(w, "  %s: %s\n", field, strings.Join(messages, ", "))
			}
		}
	default:
		fmt.Fprintf(w, "Error: %s\n", apiErr.Message)
	}
}

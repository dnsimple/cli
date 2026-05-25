package cli

import (
	"github.com/dnsimple/cli/internal/output"
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
	"github.com/spf13/cobra"
)

// pageHint builds the table pagination discovery hint from an API pagination
// response. It returns nil when pagination is absent (for example when --all
// already fetched every page), which tells the printer not to emit a hint.
func pageHint(cmd *cobra.Command, pg *dnsimple.Pagination, shown int, noun string) *output.PageInfo {
	if pg == nil {
		return nil
	}
	return &output.PageInfo{
		Noun:         noun,
		Shown:        shown,
		CurrentPage:  pg.CurrentPage,
		TotalPages:   pg.TotalPages,
		TotalEntries: pg.TotalEntries,
		CanFetchAll:  cmd.Flags().Lookup("all") != nil,
		CanPaginate:  cmd.Flags().Lookup("page") != nil,
	}
}

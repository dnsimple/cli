package cli

import (
	"github.com/dnsimple/cli/internal/cmdutil"
	"github.com/dnsimple/cli/internal/output"
	"github.com/dnsimple/cli/internal/pagination"
	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/spf13/cobra"
)

// listFlags holds the pagination flags shared by every list command.
type listFlags struct {
	page    int
	perPage int
	all     bool
}

// register adds the standard --all/--page/--per-page flags to cmd.
func (lf *listFlags) register(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&lf.all, "all", false, "Fetch all pages")
	cmd.Flags().IntVar(&lf.page, "page", 0, "Page number")
	cmd.Flags().IntVar(&lf.perPage, "per-page", 0, "Number of items per page")
}

// runList executes a paginated list command. fetch retrieves a single page for
// the given page/per-page (0 means "use the API default"); wrap adapts a page of
// results into a renderable value. With --all every page is fetched and the hint
// is skipped; otherwise the requested page is rendered with a discovery hint for
// table output.
func runList[T any, F output.Formattable](
	cmd *cobra.Command,
	f *cmdutil.Factory,
	lf *listFlags,
	noun string,
	fetch func(page, perPage int) ([]T, *dnsimple.Pagination, error),
	wrap func(items []T, pg *dnsimple.Pagination) F,
) error {
	if lf.all {
		items, err := pagination.All(func(p int) ([]T, *dnsimple.Pagination, error) {
			return fetch(p, lf.perPage)
		})
		if err != nil {
			return err
		}
		return f.Printer(cmd).Print(wrap(items, nil))
	}

	items, pg, err := fetch(lf.page, lf.perPage)
	if err != nil {
		return err
	}
	return f.Printer(cmd).PrintList(wrap(items, pg), pageHint(cmd, pg, len(items), noun))
}

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

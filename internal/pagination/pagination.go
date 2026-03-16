package pagination

import (
	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
)

// FetchPage is a function that fetches a single page of results.
type FetchPage[T any] func(page int) ([]T, *dnsimple.Pagination, error)

// All fetches all pages of results by calling fetchPage repeatedly.
func All[T any](fetchPage FetchPage[T]) ([]T, error) {
	var all []T
	page := 1

	for {
		items, pagination, err := fetchPage(page)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)

		if pagination == nil || page >= pagination.TotalPages {
			break
		}
		page++
	}

	return all, nil
}

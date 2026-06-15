package pagination

import (
	"errors"
	"testing"

	"github.com/dnsimple/dnsimple-go/v9/dnsimple"
	"github.com/stretchr/testify/assert"
)

func TestAllAggregatesAllPages(t *testing.T) {
	items, err := All(func(page int) ([]int, *dnsimple.Pagination, error) {
		switch page {
		case 1:
			return []int{1, 2}, &dnsimple.Pagination{TotalPages: 3}, nil
		case 2:
			return []int{3}, &dnsimple.Pagination{TotalPages: 3}, nil
		case 3:
			return []int{4, 5}, &dnsimple.Pagination{TotalPages: 3}, nil
		default:
			return nil, nil, errors.New("unexpected page")
		}
	})
	if !assert.NoError(t, err) {
		return
	}

	want := []int{1, 2, 3, 4, 5}
	assert.Equal(t, want, items)
}

func TestAllStopsWithoutPagination(t *testing.T) {
	calls := 0

	items, err := All(func(page int) ([]string, *dnsimple.Pagination, error) {
		calls++
		return []string{"only"}, nil, nil
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 1, calls)
	assert.Equal(t, []string{"only"}, items)
}

func TestAllPropagatesErrors(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := All(func(page int) ([]int, *dnsimple.Pagination, error) {
		if page == 2 {
			return nil, nil, wantErr
		}
		return []int{page}, &dnsimple.Pagination{TotalPages: 3}, nil
	})
	assert.ErrorIs(t, err, wantErr)
}

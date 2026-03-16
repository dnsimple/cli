package pagination

import (
	"errors"
	"testing"

	"github.com/dnsimple/dnsimple-go/v8/dnsimple"
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
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}

	want := []int{1, 2, 3, 4, 5}
	if len(items) != len(want) {
		t.Fatalf("len(items) = %d, want %d", len(items), len(want))
	}
	for i := range want {
		if items[i] != want[i] {
			t.Fatalf("items[%d] = %d, want %d", i, items[i], want[i])
		}
	}
}

func TestAllStopsWithoutPagination(t *testing.T) {
	calls := 0

	items, err := All(func(page int) ([]string, *dnsimple.Pagination, error) {
		calls++
		return []string{"only"}, nil, nil
	})
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if len(items) != 1 || items[0] != "only" {
		t.Fatalf("items = %#v, want %#v", items, []string{"only"})
	}
}

func TestAllPropagatesErrors(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := All(func(page int) ([]int, *dnsimple.Pagination, error) {
		if page == 2 {
			return nil, nil, wantErr
		}
		return []int{page}, &dnsimple.Pagination{TotalPages: 3}, nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("All() error = %v, want %v", err, wantErr)
	}
}

package handle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatUncountedPaginationHeader(t *testing.T) {
	cases := []struct {
		name        string
		page        int64
		resultCount int
		want        string
	}{
		{"empty first page", 1, 0, "Found 0 results.\n\n"},
		{"empty later page", 3, 0, "No results on page 3.\n\n"},
		{"short page is final", 2, 7, "Showing 7 results (page 2, final page).\n\n"},
		{"full page may have more", 1, 25, "Showing 25 results (page 1).\nUse page=2 to see next page.\n\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, formatUncountedPaginationHeader(tc.page, 25, tc.resultCount))
		})
	}
}

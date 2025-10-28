package freeipa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRangeFromSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		slSrc      []string
		slExpected []string
		limit      int32
		offset     int32
		total      uint32
	}{
		{
			name:       "case 1",
			slSrc:      []string{"a", "b", "c"},
			slExpected: []string{"a"},
			limit:      1,
			offset:     0,
		},
		{
			name:       "case 2",
			slSrc:      []string{"a", "b", "c"},
			slExpected: []string{"b", "c"},
			limit:      10,
			offset:     1,
		},
		{
			name:       "case 3",
			slSrc:      []string{"a", "b", "c"},
			slExpected: nil,
			limit:      0,
			offset:     10,
		},
		{
			name:       "case 4",
			slSrc:      []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}, //nolint:lll
			slExpected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20"},                   //nolint:lll
			limit:      -10,
			offset:     -10,
		},
		{
			name:       "case 5",
			slSrc:      []string{},
			slExpected: nil,
			limit:      1,
			offset:     0,
		},
		{
			name:       "case 6",
			slSrc:      nil,
			slExpected: nil,
			limit:      1,
			offset:     0,
		},
		{
			name:       "case 7",
			slSrc:      []string{"a", "b", "c"},
			slExpected: []string{"a", "b", "c"},
			limit:      3,
			offset:     0,
		},
		{
			name:       "case 8",
			slSrc:      []string{"a", "b", "c"},
			slExpected: []string{"b", "c"},
			limit:      2,
			offset:     1,
		},
		{
			name:       "case 9",
			slSrc:      []string{"a"},
			slExpected: nil,
			limit:      1,
			offset:     10,
		},
		{
			name:       "case 10",
			slSrc:      []string{"a"},
			slExpected: nil,
			limit:      100,
			offset:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.slExpected, getRangeFromSlice(tt.slSrc, tt.limit, tt.offset, limitDefault))
		})
	}
}

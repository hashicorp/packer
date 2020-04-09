package common

import "testing"

func TestStringsContains(t *testing.T) {

	tests := []struct {
		name     string
		haystack []string
		needle   string
		want     bool
	}{
		{
			name:     "found",
			haystack: []string{"a", "b", "c"},
			needle:   "b",
			want:     true,
		},
		{
			name:     "missing",
			haystack: []string{"a", "b", "c"},
			needle:   "D",
			want:     false,
		},
		{
			name:     "case insensitive",
			haystack: []string{"a", "b", "c"},
			needle:   "B",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringsContains(tt.haystack, tt.needle); got != tt.want {
				t.Errorf("StringsContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

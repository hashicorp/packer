package password

import (
	"testing"
	"unicode"
)

func TestCreateString(t *testing.T) {
	tests := []struct {
		name    string
		params  StringParams
		wantErr bool
	}{
		{
			name: "Valid params with upper, lower, numeric, and special characters",
			params: StringParams{
				Length:     12,
				Upper:      true,
				Lower:      true,
				Numeric:    true,
				Special:    true,
				MinUpper:   2,
				MinLower:   2,
				MinNumeric: 2,
				MinSpecial: 2,
			},
			wantErr: false,
		},
		{
			name: "Empty character set",
			params: StringParams{
				Length: 10,
			},
			wantErr: true,
		},
		{
			name: "Override special characters",
			params: StringParams{
				Length:          10,
				Special:         true,
				OverrideSpecial: "~!@",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CreateString(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != int(tt.params.Length) {
				t.Errorf("Expected length %d, got %d", tt.params.Length, len(result))
			}

			// Additional checks for specific character types
			if tt.params.Upper {
				count := countRunes(result, unicode.IsUpper)
				if count < int(tt.params.MinUpper) {
					t.Errorf("Expected at least %d uppercase characters, got %d", tt.params.MinUpper, count)
				}
			}
			if tt.params.Lower {
				count := countRunes(result, unicode.IsLower)
				if count < int(tt.params.MinLower) {
					t.Errorf("Expected at least %d lowercase characters, got %d", tt.params.MinLower, count)
				}
			}
			if tt.params.Numeric {
				count := countRunes(result, unicode.IsDigit)
				if count < int(tt.params.MinNumeric) {
					t.Errorf("Expected at least %d numeric characters, got %d", tt.params.MinNumeric, count)
				}
			}
		})
	}
}

func countRunes(s string, fn func(rune) bool) int {
	count := 0
	for _, r := range s {
		if fn(r) {
			count++
		}
	}
	return count
}

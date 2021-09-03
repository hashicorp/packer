package env

import (
	"os"
	"testing"
)

func Test_IsPAREnabled(t *testing.T) {
	tcs := []struct {
		name   string
		value  string
		output bool
	}{
		{
			name:   "set with 1",
			value:  "1",
			output: true,
		},
		{
			name:   "set with ON",
			value:  "ON",
			output: true,
		},
		{
			name:   "set with 0",
			value:  "0",
			output: false,
		},
		{
			name:   "set with OFF",
			value:  "OFF",
			output: false,
		},
		{
			name:   "unset",
			value:  "",
			output: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != "" {
				_ = os.Setenv(HCPPackerRegistry, tc.value)
				defer os.Unsetenv(HCPPackerRegistry)
			}
			out := IsPAREnabled()
			if out != tc.output {
				t.Fatalf("unexpected output: %t", out)
			}
		})
	}
}

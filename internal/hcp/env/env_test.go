// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package env

import (
	"testing"
)

func Test_IsHCPDisabled(t *testing.T) {
	tcs := []struct {
		name           string
		registry_value string
		output         bool
	}{
		{
			name:           "nothing set",
			registry_value: "",
			output:         false,
		},
		{
			name:           "registry set with 1",
			registry_value: "1",
			output:         false,
		},
		{
			name:           "registry set with 0",
			registry_value: "0",
			output:         true,
		},
		{
			name:           "registry set with OFF",
			registry_value: "OFF",
			output:         true,
		},
		{
			name:           "registry set with off",
			registry_value: "off",
			output:         true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(HCPPackerRegistry, tc.registry_value)
			out := IsHCPDisabled()
			if out != tc.output {
				t.Fatalf("unexpected output: %t", out)
			}
		})
	}
}

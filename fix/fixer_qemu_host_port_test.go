// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerQEMUHostPort_impl(t *testing.T) {
	var _ Fixer = new(FixerQEMUHostPort)
}

func TestFixerQEMUHostPort(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"type":              "qemu",
				"ssh_host_port_min": 2222,
			},

			Expected: map[string]any{
				"type":          "qemu",
				"host_port_min": 2222,
			},
		},
		{
			Input: map[string]any{
				"type":              "qemu",
				"ssh_host_port_max": 4444,
			},

			Expected: map[string]any{
				"type":          "qemu",
				"host_port_max": 4444,
			},
		},
	}

	for _, tc := range cases {
		var f FixerQEMUHostPort

		input := map[string]any{
			"builders": []map[string]any{tc.Input},
		}

		expected := map[string]any{
			"builders": []map[string]any{tc.Expected},
		}

		output, err := f.Fix(input)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(output, expected) {
			t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
		}
	}
}

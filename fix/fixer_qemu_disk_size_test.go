// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerQEMUDiskSize_impl(t *testing.T) {
	var _ Fixer = new(FixerQEMUDiskSize)
}

func TestFixerQEMUDiskSize(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"type":      "qemu",
				"disk_size": int(40960),
			},

			Expected: map[string]any{
				"type":      "qemu",
				"disk_size": "40960M",
			},
		},
		{
			Input: map[string]any{
				"type":      "qemu",
				"disk_size": float64(50000),
			},

			Expected: map[string]any{
				"type":      "qemu",
				"disk_size": "50000M",
			},
		},
	}

	for _, tc := range cases {
		var f FixerQEMUDiskSize

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

// Copyright (c) HashiCorp, Inc.
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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type":      "qemu",
				"disk_size": int(40960),
			},

			Expected: map[string]interface{}{
				"type":      "qemu",
				"disk_size": "40960M",
			},
		},
		{
			Input: map[string]interface{}{
				"type":      "qemu",
				"disk_size": float64(50000),
			},

			Expected: map[string]interface{}{
				"type":      "qemu",
				"disk_size": "50000M",
			},
		},
	}

	for _, tc := range cases {
		var f FixerQEMUDiskSize

		input := map[string]interface{}{
			"builders": []map[string]interface{}{tc.Input},
		}

		expected := map[string]interface{}{
			"builders": []map[string]interface{}{tc.Expected},
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

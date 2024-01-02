// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerParallelsDeprecations(t *testing.T) {
	var _ Fixer = new(FixerParallelsDeprecations)
}

func TestFixerParallelsDeprecations_Fix_parallels_tools_guest_path(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No parallels_tools_host_path field
		{
			Input: map[string]interface{}{
				"type": "parallels-iso",
			},

			Expected: map[string]interface{}{
				"type": "parallels-iso",
			},
		},

		// parallels_tools_host_path field
		{
			Input: map[string]interface{}{
				"type":                      "parallels-iso",
				"parallels_tools_host_path": "/Path...",
			},

			Expected: map[string]interface{}{
				"type": "parallels-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerParallelsDeprecations

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

func TestFixerParallelsDeprecations_Fix_guest_os_distribution(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No guest_os_distribution field
		{
			Input: map[string]interface{}{
				"type":          "parallels-iso",
				"guest_os_type": "ubuntu",
			},

			Expected: map[string]interface{}{
				"type":          "parallels-iso",
				"guest_os_type": "ubuntu",
			},
		},

		// guest_os_distribution and guest_os_type field
		{
			Input: map[string]interface{}{
				"type":                  "parallels-iso",
				"guest_os_type":         "linux",
				"guest_os_distribution": "ubuntu",
			},

			Expected: map[string]interface{}{
				"type":          "parallels-iso",
				"guest_os_type": "ubuntu",
			},
		},

		// guest_os_distribution but no guest_os_type field
		{
			Input: map[string]interface{}{
				"type":                  "parallels-iso",
				"guest_os_distribution": "ubuntu",
			},

			Expected: map[string]interface{}{
				"type":          "parallels-iso",
				"guest_os_type": "ubuntu",
			},
		},
	}

	for _, tc := range cases {
		var f FixerParallelsDeprecations

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

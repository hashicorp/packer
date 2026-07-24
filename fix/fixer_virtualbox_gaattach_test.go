// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerVirtualBoxGAAttach_Impl(t *testing.T) {
	var _ Fixer = new(FixerVirtualBoxGAAttach)
}

func TestFixerVirtualBoxGAAttach_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// No attach field
		{
			Input: map[string]any{
				"type": "virtualbox",
			},

			Expected: map[string]any{
				"type": "virtualbox",
			},
		},

		// Attach field == false
		{
			Input: map[string]any{
				"type":                   "virtualbox",
				"guest_additions_attach": false,
			},

			Expected: map[string]any{
				"type":                 "virtualbox",
				"guest_additions_mode": "upload",
			},
		},

		// Attach field == true
		{
			Input: map[string]any{
				"type":                   "virtualbox",
				"guest_additions_attach": true,
			},

			Expected: map[string]any{
				"type":                 "virtualbox",
				"guest_additions_mode": "attach",
			},
		},

		// Attach field is not a bool
		{
			Input: map[string]any{
				"type":                   "virtualbox",
				"guest_additions_attach": "what",
			},

			Expected: map[string]any{
				"type":                   "virtualbox",
				"guest_additions_attach": "what",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxGAAttach

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

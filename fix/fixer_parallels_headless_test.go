// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerParallelsHeadless_Impl(t *testing.T) {
	var _ Fixer = new(FixerParallelsHeadless)
}

func TestFixerParallelsHeadless_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// No headless field
		{
			Input: map[string]any{
				"type": "parallels-iso",
			},

			Expected: map[string]any{
				"type": "parallels-iso",
			},
		},

		// Headless field
		{
			Input: map[string]any{
				"type":     "parallels-iso",
				"headless": false,
			},

			Expected: map[string]any{
				"type": "parallels-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerParallelsHeadless

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

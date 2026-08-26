// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerAzureExcludeFromLatest(t *testing.T) {
	var _ Fixer = new(FixerAzureExcludeFromLatest)
}

func TestFixerAzureExcludeFromLatest_Fix_exlude_from_latest(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// No shared_image_destination field
		{
			Input: map[string]any{
				"type": "azure-chroot",
			},

			Expected: map[string]any{
				"type": "azure-chroot",
			},
		},

		// exlude_from_latest field
		{
			Input: map[string]any{
				"type": "azure-chroot",
				"shared_image_destination": map[string]any{
					"exlude_from_latest": "false",
				},
			},

			Expected: map[string]any{
				"type": "azure-chroot",
				"shared_image_destination": map[string]any{
					"exclude_from_latest": "false",
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerAzureExcludeFromLatest

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

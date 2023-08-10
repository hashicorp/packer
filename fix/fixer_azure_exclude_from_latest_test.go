// Copyright (c) HashiCorp, Inc.
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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No shared_image_destination field
		{
			Input: map[string]interface{}{
				"type": "azure-chroot",
			},

			Expected: map[string]interface{}{
				"type": "azure-chroot",
			},
		},

		// exlude_from_latest field
		{
			Input: map[string]interface{}{
				"type": "azure-chroot",
				"shared_image_destination": map[string]interface{}{
					"exlude_from_latest": "false",
				},
			},

			Expected: map[string]interface{}{
				"type": "azure-chroot",
				"shared_image_destination": map[string]interface{}{
					"exclude_from_latest": "false",
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerAzureExcludeFromLatest

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

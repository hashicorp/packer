// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerAmazonEnhancedNetworking_Impl(t *testing.T) {
	var _ Fixer = new(FixerAmazonEnhancedNetworking)
}

func TestFixerAmazonEnhancedNetworking(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// Attach field == false
		{
			Input: map[string]any{
				"type":                "amazon-ebs",
				"enhanced_networking": false,
			},

			Expected: map[string]any{
				"type":        "amazon-ebs",
				"ena_support": false,
			},
		},

		// Attach field == true
		{
			Input: map[string]any{
				"type":                "amazon-ebs",
				"enhanced_networking": true,
			},

			Expected: map[string]any{
				"type":        "amazon-ebs",
				"ena_support": true,
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonEnhancedNetworking

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

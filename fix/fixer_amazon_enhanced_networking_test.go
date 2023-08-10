// Copyright (c) HashiCorp, Inc.
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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// Attach field == false
		{
			Input: map[string]interface{}{
				"type":                "amazon-ebs",
				"enhanced_networking": false,
			},

			Expected: map[string]interface{}{
				"type":        "amazon-ebs",
				"ena_support": false,
			},
		},

		// Attach field == true
		{
			Input: map[string]interface{}{
				"type":                "amazon-ebs",
				"enhanced_networking": true,
			},

			Expected: map[string]interface{}{
				"type":        "amazon-ebs",
				"ena_support": true,
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonEnhancedNetworking

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

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerAmazonShutdownBehavior(t *testing.T) {
	var _ Fixer = new(FixerAmazonShutdownBehavior)
}

func TestFixerAmazonShutdownBehavior_Fix_shutdown_behaviour(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No shutdown_behaviour field
		{
			Input: map[string]interface{}{
				"type": "amazon-ebs",
			},

			Expected: map[string]interface{}{
				"type": "amazon-ebs",
			},
		},

		// shutdown_behaviour field
		{
			Input: map[string]interface{}{
				"type":               "amazon-ebs",
				"shutdown_behaviour": "stop",
			},

			Expected: map[string]interface{}{
				"type":              "amazon-ebs",
				"shutdown_behavior": "stop",
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonShutdownBehavior

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

// Copyright IBM Corp. 2024, 2025
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
		Input    map[string]any
		Expected map[string]any
	}{
		// No shutdown_behaviour field
		{
			Input: map[string]any{
				"type": "amazon-ebs",
			},

			Expected: map[string]any{
				"type": "amazon-ebs",
			},
		},

		// shutdown_behaviour field
		{
			Input: map[string]any{
				"type":               "amazon-ebs",
				"shutdown_behaviour": "stop",
			},

			Expected: map[string]any{
				"type":              "amazon-ebs",
				"shutdown_behavior": "stop",
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonShutdownBehavior

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

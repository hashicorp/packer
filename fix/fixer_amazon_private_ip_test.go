// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerAmazonPrivateIP_Impl(t *testing.T) {
	var _ Fixer = new(FixerAmazonPrivateIP)
}

func TestFixerAmazonPrivateIP(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// Attach field == false
		{
			Input: map[string]any{
				"type":           "amazon-ebs",
				"ssh_private_ip": false,
			},

			Expected: map[string]any{
				"type":          "amazon-ebs",
				"ssh_interface": "public_ip",
			},
		},

		// Attach field == true
		{
			Input: map[string]any{
				"type":           "amazon-ebs",
				"ssh_private_ip": true,
			},

			Expected: map[string]any{
				"type":          "amazon-ebs",
				"ssh_interface": "private_ip",
			},
		},

		// ssh_private_ip specified as string
		{
			Input: map[string]any{
				"type":           "amazon-ebs",
				"ssh_private_ip": "true",
			},

			Expected: map[string]any{
				"type":          "amazon-ebs",
				"ssh_interface": "private_ip",
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonPrivateIP

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

func TestFixerAmazonPrivateIPNonBoolean(t *testing.T) {
	var f FixerAmazonPrivateIP

	input := map[string]any{
		"builders": []map[string]any{{
			"type":           "amazon-ebs",
			"ssh_private_ip": "not-a-boolean-value",
		}},
	}

	_, err := f.Fix(input)
	if err == nil {
		t.Fatal("should have errored")
	}
}

// Copyright (c) HashiCorp, Inc.
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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// Attach field == false
		{
			Input: map[string]interface{}{
				"type":           "amazon-ebs",
				"ssh_private_ip": false,
			},

			Expected: map[string]interface{}{
				"type":          "amazon-ebs",
				"ssh_interface": "public_ip",
			},
		},

		// Attach field == true
		{
			Input: map[string]interface{}{
				"type":           "amazon-ebs",
				"ssh_private_ip": true,
			},

			Expected: map[string]interface{}{
				"type":          "amazon-ebs",
				"ssh_interface": "private_ip",
			},
		},

		// ssh_private_ip specified as string
		{
			Input: map[string]interface{}{
				"type":           "amazon-ebs",
				"ssh_private_ip": "true",
			},

			Expected: map[string]interface{}{
				"type":          "amazon-ebs",
				"ssh_interface": "private_ip",
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonPrivateIP

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

func TestFixerAmazonPrivateIPNonBoolean(t *testing.T) {
	var f FixerAmazonPrivateIP

	input := map[string]interface{}{
		"builders": []map[string]interface{}{{
			"type":           "amazon-ebs",
			"ssh_private_ip": "not-a-boolean-value",
		}},
	}

	_, err := f.Fix(input)
	if err == nil {
		t.Fatal("should have errored")
	}
}

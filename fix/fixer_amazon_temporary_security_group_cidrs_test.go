// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerAmazonTemporarySecurityCIDRs_Impl(t *testing.T) {
	var _ Fixer = new(FixerAmazonTemporarySecurityCIDRs)
}

func TestFixerAmazonTemporarySecurityCIDRs(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type":                                 "amazon-ebs",
				"temporary_security_group_source_cidr": "0.0.0.0/0",
			},

			Expected: map[string]interface{}{
				"type":                                  "amazon-ebs",
				"temporary_security_group_source_cidrs": []string{"0.0.0.0/0"},
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonTemporarySecurityCIDRs

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

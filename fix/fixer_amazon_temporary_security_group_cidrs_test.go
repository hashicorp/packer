// Copyright IBM Corp. 2024, 2025
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
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"type":                                 "amazon-ebs",
				"temporary_security_group_source_cidr": "0.0.0.0/0",
			},

			Expected: map[string]any{
				"type":                                  "amazon-ebs",
				"temporary_security_group_source_cidrs": []string{"0.0.0.0/0"},
			},
		},
	}

	for _, tc := range cases {
		var f FixerAmazonTemporarySecurityCIDRs

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

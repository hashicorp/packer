// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerScalewayAccessKey_Fix_Impl(t *testing.T) {
	var _ Fixer = new(FixerScalewayAccessKey)
}

func TestFixerScalewayAccessKey_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No key_path field
		{
			Input: map[string]interface{}{
				"type": "scaleway",
			},

			Expected: map[string]interface{}{
				"type": "scaleway",
			},
		},

		// organization_id without access_key
		{
			Input: map[string]interface{}{
				"type":            "scaleway",
				"organization_id": "0000",
			},

			Expected: map[string]interface{}{
				"type":            "scaleway",
				"organization_id": "0000",
			},
		},

		// access_key without organization_id
		{
			Input: map[string]interface{}{
				"type":       "scaleway",
				"access_key": "1111",
			},

			Expected: map[string]interface{}{
				"type":            "scaleway",
				"organization_id": "1111",
			},
		},

		// access_key and organization_id
		{
			Input: map[string]interface{}{
				"type":            "scaleway",
				"access_key":      "2222",
				"organization_id": "3333",
			},

			Expected: map[string]interface{}{
				"type":            "scaleway",
				"organization_id": "3333",
			},
		},
	}

	for _, tc := range cases {
		var f FixerScalewayAccessKey

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

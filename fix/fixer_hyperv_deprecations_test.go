// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerHypervDeprecations_impl(t *testing.T) {
	var _ Fixer = new(FixerHypervDeprecations)
}

func TestFixerHypervDeprecations_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No vhd_temp_path field in template - noop
		{
			Input: map[string]interface{}{
				"type": "hyperv-iso",
			},

			Expected: map[string]interface{}{
				"type": "hyperv-iso",
			},
		},

		// Deprecated vhd_temp_path field in template should be deleted
		{
			Input: map[string]interface{}{
				"type":          "hyperv-iso",
				"vhd_temp_path": "foopath",
			},

			Expected: map[string]interface{}{
				"type": "hyperv-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerHypervDeprecations

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

		assert.Equal(t, expected, output, "Should be equal")
	}
}

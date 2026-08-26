// Copyright IBM Corp. 2024, 2025
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
		Input    map[string]any
		Expected map[string]any
	}{
		// No vhd_temp_path field in template - noop
		{
			Input: map[string]any{
				"type": "hyperv-iso",
			},

			Expected: map[string]any{
				"type": "hyperv-iso",
			},
		},

		// Deprecated vhd_temp_path field in template should be deleted
		{
			Input: map[string]any{
				"type":          "hyperv-iso",
				"vhd_temp_path": "foopath",
			},

			Expected: map[string]any{
				"type": "hyperv-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerHypervDeprecations

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

		assert.Equal(t, expected, output, "Should be equal")
	}
}

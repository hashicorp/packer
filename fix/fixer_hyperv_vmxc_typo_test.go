// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerHypervVmxcTypo_impl(t *testing.T) {
	var _ Fixer = new(FixerHypervVmxcTypo)
}

func TestFixerHypervVmxcTypo_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		// No "clone_from_vmxc_path" in template - noop
		{
			Input: map[string]any{
				"type":      "hyperv-vmcx",
				"temp_path": "C:/some/temp/path",
			},

			Expected: map[string]any{
				"type":      "hyperv-vmcx",
				"temp_path": "C:/some/temp/path",
			},
		},

		// "clone_from_vmxc_path" should be replaced with
		// "clone_from_vmcx_path" in template
		{
			Input: map[string]any{
				"type":                 "hyperv-vmcx",
				"clone_from_vmxc_path": "C:/some/vmcx/path",
			},

			Expected: map[string]any{
				"type":                 "hyperv-vmcx",
				"clone_from_vmcx_path": "C:/some/vmcx/path",
			},
		},
	}

	for _, tc := range cases {
		var f FixerHypervVmxcTypo

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

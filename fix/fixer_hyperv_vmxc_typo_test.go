// Copyright (c) HashiCorp, Inc.
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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No "clone_from_vmxc_path" in template - noop
		{
			Input: map[string]interface{}{
				"type":      "hyperv-vmcx",
				"temp_path": "C:/some/temp/path",
			},

			Expected: map[string]interface{}{
				"type":      "hyperv-vmcx",
				"temp_path": "C:/some/temp/path",
			},
		},

		// "clone_from_vmxc_path" should be replaced with
		// "clone_from_vmcx_path" in template
		{
			Input: map[string]interface{}{
				"type":                 "hyperv-vmcx",
				"clone_from_vmxc_path": "C:/some/vmcx/path",
			},

			Expected: map[string]interface{}{
				"type":                 "hyperv-vmcx",
				"clone_from_vmcx_path": "C:/some/vmcx/path",
			},
		},
	}

	for _, tc := range cases {
		var f FixerHypervVmxcTypo

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

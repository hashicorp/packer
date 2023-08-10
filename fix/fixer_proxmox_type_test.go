// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerProxmoxType_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerProxmoxType)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerProxmoxType_Fix(t *testing.T) {

	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{

		{
			Input: map[string]interface{}{
				"type": "proxmox",
			},

			Expected: map[string]interface{}{
				"type": "proxmox-iso",
			},
		},

		{
			Input: map[string]interface{}{
				"type": "proxmox-iso",
			},

			Expected: map[string]interface{}{
				"type": "proxmox-iso",
			},
		},

		{
			Input: map[string]interface{}{
				"type": "proxmox-clone",
			},

			Expected: map[string]interface{}{
				"type": "proxmox-clone",
			},
		},
	}

	for _, tc := range cases {
		var f FixerProxmoxType

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

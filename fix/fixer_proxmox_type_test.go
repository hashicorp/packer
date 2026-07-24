// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerProxmoxType_Impl(t *testing.T) {
	var raw any = new(FixerProxmoxType)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerProxmoxType_Fix(t *testing.T) {

	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{

		{
			Input: map[string]any{
				"type": "proxmox",
			},

			Expected: map[string]any{
				"type": "proxmox-iso",
			},
		},

		{
			Input: map[string]any{
				"type": "proxmox-iso",
			},

			Expected: map[string]any{
				"type": "proxmox-iso",
			},
		},

		{
			Input: map[string]any{
				"type": "proxmox-clone",
			},

			Expected: map[string]any{
				"type": "proxmox-clone",
			},
		},
	}

	for _, tc := range cases {
		var f FixerProxmoxType

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

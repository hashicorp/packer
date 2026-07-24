// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerVagrantPPOverride_Impl(t *testing.T) {
	var _ Fixer = new(FixerVagrantPPOverride)
}

func TestFixerVagrantPPOverride_Fix(t *testing.T) {
	var f FixerVagrantPPOverride

	input := map[string]any{
		"post-processors": []any{
			"foo",

			map[string]any{
				"type": "vagrant",
				"aws": map[string]any{
					"foo": "bar",
				},
			},

			map[string]any{
				"type": "vsphere",
			},

			[]any{
				map[string]any{
					"type": "vagrant",
					"vmware": map[string]any{
						"foo": "bar",
					},
				},
			},
		},
	}

	expected := map[string]any{
		"post-processors": []any{
			"foo",

			map[string]any{
				"type": "vagrant",
				"override": map[string]any{
					"aws": map[string]any{
						"foo": "bar",
					},
				},
			},

			map[string]any{
				"type": "vsphere",
			},

			[]any{
				map[string]any{
					"type": "vagrant",
					"override": map[string]any{
						"vmware": map[string]any{
							"foo": "bar",
						},
					},
				},
			},
		},
	}

	output, err := f.Fix(input)
	assert.NoError(t, err)

	assert.Equal(t, expected, output)
}

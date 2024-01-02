// Copyright (c) HashiCorp, Inc.
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

	input := map[string]interface{}{
		"post-processors": []interface{}{
			"foo",

			map[string]interface{}{
				"type": "vagrant",
				"aws": map[string]interface{}{
					"foo": "bar",
				},
			},

			map[string]interface{}{
				"type": "vsphere",
			},

			[]interface{}{
				map[string]interface{}{
					"type": "vagrant",
					"vmware": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	}

	expected := map[string]interface{}{
		"post-processors": []interface{}{
			"foo",

			map[string]interface{}{
				"type": "vagrant",
				"override": map[string]interface{}{
					"aws": map[string]interface{}{
						"foo": "bar",
					},
				},
			},

			map[string]interface{}{
				"type": "vsphere",
			},

			[]interface{}{
				map[string]interface{}{
					"type": "vagrant",
					"override": map[string]interface{}{
						"vmware": map[string]interface{}{
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

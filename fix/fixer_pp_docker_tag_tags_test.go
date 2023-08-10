// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerDockerTags(t *testing.T) {
	var _ Fixer = new(FixerVagrantPPOverride)
}

func TestFixerDockerTags_Fix(t *testing.T) {
	var f FixerDockerTagtoTags

	input := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type": "docker-tag",
				"tag":  "foo",
				"tags": []string{"foo", "bar"},
			},
			[]interface{}{
				map[string]interface{}{
					"type": "docker-tag",
					"tag":  []string{"baz"},
					"tags": []string{"foo", "bar"},
				},
			},
		},
	}

	expected := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type": "docker-tag",
				"tags": []string{"foo", "bar"},
			},
			[]interface{}{
				map[string]interface{}{
					"type": "docker-tag",
					"tags": []string{"baz", "foo", "bar"},
				},
			},
		},
	}

	output, err := f.Fix(input)
	assert.NoError(t, err)
	for k, v := range expected {
		assert.EqualValues(t, v, output[k])
	}
}

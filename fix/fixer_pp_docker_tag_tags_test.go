// Copyright IBM Corp. 2024, 2025
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

	input := map[string]any{
		"post-processors": []any{
			map[string]any{
				"type": "docker-tag",
				"tag":  "foo",
				"tags": []string{"foo", "bar"},
			},
			[]any{
				map[string]any{
					"type": "docker-tag",
					"tag":  []string{"baz"},
					"tags": []string{"foo", "bar"},
				},
			},
		},
	}

	expected := map[string]any{
		"post-processors": []any{
			map[string]any{
				"type": "docker-tag",
				"tags": []string{"foo", "bar"},
			},
			[]any{
				map[string]any{
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

// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerManifestPPFilename_Impl(t *testing.T) {
	var _ Fixer = new(FixerVagrantPPOverride)
}

func TestFixerManifestPPFilename_Fix(t *testing.T) {
	var f FixerManifestFilename

	input := map[string]any{
		"post-processors": []any{
			map[string]any{
				"type":     "manifest",
				"filename": "foo",
			},
			[]any{
				map[string]any{
					"type":     "manifest",
					"filename": "foo",
				},
			},
		},
	}

	expected := map[string]any{
		"post-processors": []any{
			map[string]any{
				"type":   "manifest",
				"output": "foo",
			},
			[]any{
				map[string]any{
					"type":   "manifest",
					"output": "foo",
				},
			},
		},
	}

	output, err := f.Fix(input)
	assert.NoError(t, err)

	assert.Equal(t, expected, output)
}

// Copyright (c) HashiCorp, Inc.
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

	input := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type":     "manifest",
				"filename": "foo",
			},
			[]interface{}{
				map[string]interface{}{
					"type":     "manifest",
					"filename": "foo",
				},
			},
		},
	}

	expected := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type":   "manifest",
				"output": "foo",
			},
			[]interface{}{
				map[string]interface{}{
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

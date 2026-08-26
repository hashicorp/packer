// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerISOChecksumTypeAndURL_Impl(t *testing.T) {
	var raw any = new(FixerISOChecksumTypeAndURL)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerISOChecksumTypeAndURL_Fix(t *testing.T) {

	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{

		{
			Input: map[string]any{
				"type":              "foo",
				"iso_checksum_url":  "bar",
				"iso_checksum_type": "ignored",
			},

			Expected: map[string]any{
				"type":         "foo",
				"iso_checksum": "file:bar",
			},
		},

		{
			Input: map[string]any{
				"type":              "foo",
				"iso_checksum":      "checksum",
				"iso_checksum_type": "md5",
			},

			Expected: map[string]any{
				"type":         "foo",
				"iso_checksum": "md5:checksum",
			},
		},

		{
			Input: map[string]any{
				"type":             "foo",
				"iso_checksum":     "checksum",
				"iso_checksum_url": "path/to/checksumfile",
			},

			Expected: map[string]any{
				"type":         "foo",
				"iso_checksum": "file:path/to/checksumfile",
			},
		},
	}

	for _, tc := range cases {
		var f FixerISOChecksumTypeAndURL

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

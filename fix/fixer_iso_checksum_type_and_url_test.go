// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixerISOChecksumTypeAndURL_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerISOChecksumTypeAndURL)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerISOChecksumTypeAndURL_Fix(t *testing.T) {

	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{

		{
			Input: map[string]interface{}{
				"type":              "foo",
				"iso_checksum_url":  "bar",
				"iso_checksum_type": "ignored",
			},

			Expected: map[string]interface{}{
				"type":         "foo",
				"iso_checksum": "file:bar",
			},
		},

		{
			Input: map[string]interface{}{
				"type":              "foo",
				"iso_checksum":      "checksum",
				"iso_checksum_type": "md5",
			},

			Expected: map[string]interface{}{
				"type":         "foo",
				"iso_checksum": "md5:checksum",
			},
		},

		{
			Input: map[string]interface{}{
				"type":             "foo",
				"iso_checksum":     "checksum",
				"iso_checksum_url": "path/to/checksumfile",
			},

			Expected: map[string]interface{}{
				"type":         "foo",
				"iso_checksum": "file:path/to/checksumfile",
			},
		},
	}

	for _, tc := range cases {
		var f FixerISOChecksumTypeAndURL

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

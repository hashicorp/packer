// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerISOMD5_Impl(t *testing.T) {
	var raw any = new(FixerISOMD5)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerISOMD5_Fix(t *testing.T) {
	var f FixerISOMD5

	input := map[string]any{
		"builders": []any{
			map[string]string{
				"type":    "foo",
				"iso_md5": "bar",
			},
		},
	}

	expected := map[string]any{
		"builders": []map[string]any{
			{
				"type":              "foo",
				"iso_checksum":      "bar",
				"iso_checksum_type": "md5",
			},
		},
	}

	output, err := f.Fix(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
	}
}

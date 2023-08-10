// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerISOMD5_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerISOMD5)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerISOMD5_Fix(t *testing.T) {
	var f FixerISOMD5

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]string{
				"type":    "foo",
				"iso_md5": "bar",
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
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

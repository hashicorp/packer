// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerCreateTime_Impl(t *testing.T) {
	var raw any = new(FixerCreateTime)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerCreateTime_Fix(t *testing.T) {
	var f FixerCreateTime

	input := map[string]any{
		"builders": []any{
			map[string]string{
				"type":     "foo",
				"ami_name": "{{.CreateTime}} foo",
			},
		},
	}

	expected := map[string]any{
		"builders": []map[string]any{
			{
				"type":     "foo",
				"ami_name": "{{timestamp}} foo",
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

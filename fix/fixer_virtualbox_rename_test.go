// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerVirtualBoxRename_impl(t *testing.T) {
	var _ Fixer = new(FixerVirtualBoxRename)
}

func TestFixerVirtualBoxRename_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"type": "virtualbox",
			},

			Expected: map[string]any{
				"type": "virtualbox-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxRename

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

		if !reflect.DeepEqual(output, expected) {
			t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
		}
	}
}

func TestFixerVirtualBoxRenameFix_provisionerOverride(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"provisioners": []any{
					map[string]any{
						"override": map[string]any{
							"virtualbox": map[string]any{},
						},
					},
				},
			},

			Expected: map[string]any{
				"provisioners": []any{
					map[string]any{
						"override": map[string]any{
							"virtualbox-iso": map[string]any{},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxRename

		output, err := f.Fix(tc.Input)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(output, tc.Expected) {
			t.Fatalf("unexpected:\n\n%#v\nexpected:\n\n%#v\n", output, tc.Expected)
		}
	}
}

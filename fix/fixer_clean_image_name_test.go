// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFixerCleanImageName_Impl(t *testing.T) {
	var raw any = new(FixerCleanImageName)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerCleanImageName_Fix(t *testing.T) {
	var f FixerCleanImageName

	input := map[string]any{
		"builders": []any{
			map[string]any{
				"type":     "foo",
				"ami_name": "heyo clean_image_name",
				"image_labels": map[string]any{
					"name": "test-packer-{{packer_version | clean_image_name}}",
				},
			},
		},
	}

	expected := map[string]any{
		"builders": []map[string]any{
			{
				"type":     "foo",
				"ami_name": "heyo clean_resource_name",
				"image_labels": map[string]any{
					"name": "test-packer-{{packer_version | clean_resource_name}}",
				},
			},
		},
	}

	output, err := f.Fix(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if diff := cmp.Diff(expected, output); diff != "" {
		t.Fatalf("unexpected output: %s", diff)
	}
}

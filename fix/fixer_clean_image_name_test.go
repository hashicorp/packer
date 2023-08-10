// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFixerCleanImageName_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerCleanImageName)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerCleanImageName_Fix(t *testing.T) {
	var f FixerCleanImageName

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]interface{}{
				"type":     "foo",
				"ami_name": "heyo clean_image_name",
				"image_labels": map[string]interface{}{
					"name": "test-packer-{{packer_version | clean_image_name}}",
				},
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
			{
				"type":     "foo",
				"ami_name": "heyo clean_resource_name",
				"image_labels": map[string]interface{}{
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

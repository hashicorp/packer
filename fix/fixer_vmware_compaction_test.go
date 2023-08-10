// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerVMwareCompaction_impl(t *testing.T) {
	var _ Fixer = new(FixerVMwareCompaction)
}

func TestFixerVMwareCompaction_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type": "virtualbox-iso",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox-iso",
			},
		},
		{
			Input: map[string]interface{}{
				"type": "vmware-iso",
			},

			Expected: map[string]interface{}{
				"type": "vmware-iso",
			},
		},
		{
			Input: map[string]interface{}{
				"type":        "vmware-iso",
				"remote_type": "esx5",
			},

			Expected: map[string]interface{}{
				"type":            "vmware-iso",
				"remote_type":     "esx5",
				"disk_type_id":    "zeroedthick",
				"skip_compaction": true,
			},
		},
		{
			Input: map[string]interface{}{
				"type":         "vmware-iso",
				"remote_type":  "esx5",
				"disk_type_id": "zeroedthick",
			},

			Expected: map[string]interface{}{
				"type":            "vmware-iso",
				"remote_type":     "esx5",
				"disk_type_id":    "zeroedthick",
				"skip_compaction": true,
			},
		},
		{
			Input: map[string]interface{}{
				"type":            "vmware-iso",
				"remote_type":     "esx5",
				"disk_type_id":    "zeroedthick",
				"skip_compaction": false,
			},

			Expected: map[string]interface{}{
				"type":            "vmware-iso",
				"remote_type":     "esx5",
				"disk_type_id":    "zeroedthick",
				"skip_compaction": true,
			},
		},
		{
			Input: map[string]interface{}{
				"type":         "vmware-iso",
				"remote_type":  "esx5",
				"disk_type_id": "thin",
			},

			Expected: map[string]interface{}{
				"type":         "vmware-iso",
				"remote_type":  "esx5",
				"disk_type_id": "thin",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVMwareCompaction

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

		if !reflect.DeepEqual(output, expected) {
			t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
		}
	}
}

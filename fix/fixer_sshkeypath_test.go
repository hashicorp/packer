// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerSSHKeyPath_Impl(t *testing.T) {
	var _ Fixer = new(FixerSSHKeyPath)
}

func TestFixerSSHKeyPath_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No key_path field
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox",
			},
		},

		// private_key_file without key_path
		{
			Input: map[string]interface{}{
				"ssh_private_key_file": "id_rsa",
			},

			Expected: map[string]interface{}{
				"ssh_private_key_file": "id_rsa",
			},
		},

		// key_path without private_key_file
		{
			Input: map[string]interface{}{
				"ssh_key_path": "id_rsa",
			},

			Expected: map[string]interface{}{
				"ssh_private_key_file": "id_rsa",
			},
		},

		// key_path and private_key_file
		{
			Input: map[string]interface{}{
				"ssh_key_path":         "key_id_rsa",
				"ssh_private_key_file": "private_id_rsa",
			},

			Expected: map[string]interface{}{
				"ssh_private_key_file": "private_id_rsa",
			},
		},
	}

	for _, tc := range cases {
		var f FixerSSHKeyPath

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

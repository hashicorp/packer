// Copyright IBM Corp. 2024, 2025
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
		Input    map[string]any
		Expected map[string]any
	}{
		// No key_path field
		{
			Input: map[string]any{
				"type": "virtualbox",
			},

			Expected: map[string]any{
				"type": "virtualbox",
			},
		},

		// private_key_file without key_path
		{
			Input: map[string]any{
				"ssh_private_key_file": "id_rsa",
			},

			Expected: map[string]any{
				"ssh_private_key_file": "id_rsa",
			},
		},

		// key_path without private_key_file
		{
			Input: map[string]any{
				"ssh_key_path": "id_rsa",
			},

			Expected: map[string]any{
				"ssh_private_key_file": "id_rsa",
			},
		},

		// key_path and private_key_file
		{
			Input: map[string]any{
				"ssh_key_path":         "key_id_rsa",
				"ssh_private_key_file": "private_id_rsa",
			},

			Expected: map[string]any{
				"ssh_private_key_file": "private_id_rsa",
			},
		},
	}

	for _, tc := range cases {
		var f FixerSSHKeyPath

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

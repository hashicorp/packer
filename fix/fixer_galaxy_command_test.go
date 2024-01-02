// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerGalaxyCommand_Impl(t *testing.T) {
	var _ Fixer = new(FixerGalaxyCommand)
}

func TestFixerGalaxyCommand_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// set galaxy_command
		{
			Input: map[string]interface{}{
				"type":           "ansible-local",
				"galaxy_command": "/usr/local/bin/ansible-galaxy",
			},

			Expected: map[string]interface{}{
				"type":           "ansible-local",
				"galaxy_command": "/usr/local/bin/ansible-galaxy",
			},
		},

		// set galaxycommand (old key)
		{
			Input: map[string]interface{}{
				"type":          "ansible-local",
				"galaxycommand": "/usr/bin/ansible-galaxy",
			},

			Expected: map[string]interface{}{
				"type":           "ansible-local",
				"galaxy_command": "/usr/bin/ansible-galaxy",
			},
		},

		// set galaxy_command and galaxycommand
		// galaxy_command takes precedence
		{
			Input: map[string]interface{}{
				"type":           "ansible-local",
				"galaxy_command": "ansible_galaxy_command",
				"galaxycommand":  "ansible_galaxycommand",
			},

			Expected: map[string]interface{}{
				"type":           "ansible-local",
				"galaxy_command": "ansible_galaxy_command",
			},
		},
	}

	for _, tc := range cases {
		var f FixerGalaxyCommand

		input := map[string]interface{}{
			"provisioners": []interface{}{tc.Input},
		}

		expected := map[string]interface{}{
			"provisioners": []interface{}{tc.Expected},
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

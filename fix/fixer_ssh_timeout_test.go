// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerSSHTimout_Impl(t *testing.T) {
	var _ Fixer = new(FixerSSHTimout)
}

func TestFixerSSHTimout_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// set galaxy_command
		{
			Input: map[string]interface{}{
				"ssh_timeout": "1h5m2s",
			},

			Expected: map[string]interface{}{
				"ssh_timeout": "1h5m2s",
			},
		},

		// set galaxycommand (old key)
		{
			Input: map[string]interface{}{
				"ssh_wait_timeout": "1h5m2s",
			},

			Expected: map[string]interface{}{
				"ssh_timeout": "1h5m2s",
			},
		},

		// set galaxy_command and galaxycommand
		// galaxy_command takes precedence
		{
			Input: map[string]interface{}{
				"ssh_timeout":      "1h5m2s",
				"ssh_wait_timeout": "30m",
			},

			Expected: map[string]interface{}{
				"ssh_timeout": "1h5m2s",
			},
		},
	}

	for _, tc := range cases {
		var f FixerSSHTimout

		input := map[string]interface{}{
			"builders": []interface{}{tc.Input},
		}

		expected := map[string]interface{}{
			"builders": []interface{}{tc.Expected},
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

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerQEMUHostPort_impl(t *testing.T) {
	var _ Fixer = new(FixerQEMUHostPort)
}

func TestFixerQEMUHostPort(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type":              "qemu",
				"ssh_host_port_min": 2222,
			},

			Expected: map[string]interface{}{
				"type":          "qemu",
				"host_port_min": 2222,
			},
		},
		{
			Input: map[string]interface{}{
				"type":              "qemu",
				"ssh_host_port_max": 4444,
			},

			Expected: map[string]interface{}{
				"type":          "qemu",
				"host_port_max": 4444,
			},
		},
	}

	for _, tc := range cases {
		var f FixerQEMUHostPort

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

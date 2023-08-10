// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerCommConfig_Impl(t *testing.T) {
	var _ Fixer = new(FixerCommConfig)
}

func TestFixerCommConfig_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// set host_port_min
		{
			Input: map[string]interface{}{
				"type":          "virtualbox-iso",
				"host_port_min": 2222,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-iso",
				"host_port_min": 2222,
			},
		},

		// set ssh_host_port_min (old key)
		{
			Input: map[string]interface{}{
				"type":              "virtualbox-ovf",
				"ssh_host_port_min": 2222,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-ovf",
				"host_port_min": 2222,
			},
		},

		// set ssh_host_port_min and host_port_min
		// host_port_min takes precedence
		{
			Input: map[string]interface{}{
				"type":              "virtualbox-vm",
				"ssh_host_port_min": 1234,
				"host_port_min":     4321,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-vm",
				"host_port_min": 4321,
			},
		},

		// set host_port_max
		{
			Input: map[string]interface{}{
				"type":          "virtualbox-iso",
				"host_port_max": 4444,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-iso",
				"host_port_max": 4444,
			},
		},

		// set ssh_host_port_max (old key)
		{
			Input: map[string]interface{}{
				"type":              "virtualbox-iso",
				"ssh_host_port_max": 4444,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-iso",
				"host_port_max": 4444,
			},
		},

		// set ssh_host_port_max and host_port_max
		// host_port_max takes precedence
		{
			Input: map[string]interface{}{
				"type":              "virtualbox-vm",
				"ssh_host_port_max": 1234,
				"host_port_max":     4321,
			},

			Expected: map[string]interface{}{
				"type":          "virtualbox-vm",
				"host_port_max": 4321,
			},
		},

		// set skip_nat_mapping
		{
			Input: map[string]interface{}{
				"type":             "virtualbox-vm",
				"skip_nat_mapping": true,
			},

			Expected: map[string]interface{}{
				"type":             "virtualbox-vm",
				"skip_nat_mapping": true,
			},
		},

		// set ssh_skip_nat_mapping (old key)
		{
			Input: map[string]interface{}{
				"type":                 "virtualbox-vm",
				"ssh_skip_nat_mapping": true,
			},

			Expected: map[string]interface{}{
				"type":             "virtualbox-vm",
				"skip_nat_mapping": true,
			},
		},

		// set ssh_skip_nat_mapping and skip_nat_mapping
		// skip_nat_mapping takes precedence
		{
			Input: map[string]interface{}{
				"type":                 "virtualbox-iso",
				"ssh_skip_nat_mapping": false,
				"skip_nat_mapping":     true,
			},

			Expected: map[string]interface{}{
				"type":             "virtualbox-iso",
				"skip_nat_mapping": true,
			},
		},
	}

	for _, tc := range cases {
		var f FixerCommConfig

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

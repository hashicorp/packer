// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerVSphereNetwork_impl(t *testing.T) {
	var _ Fixer = new(FixerVSphereNetworkDisk)
}

func TestFixerVSphereNetwork_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]any
		Expected map[string]any
	}{
		{
			Input: map[string]any{
				"type":        "vsphere-iso",
				"network":     "",
				"networkCard": "vmxnet3",
				"disk_size":   5000,
			},

			Expected: map[string]any{
				"type": "vsphere-iso",
				"network_adapters": []any{
					map[string]any{
						"network":      "",
						"network_card": "vmxnet3",
					},
				},
				"storage": []any{
					map[string]any{
						"disk_size": 5000,
					},
				},
			},
		},
		{
			Input: map[string]any{
				"type":         "vsphere-iso",
				"network":      "",
				"network_card": "vmxnet3",
				"disk_size":    5000,
			},

			Expected: map[string]any{
				"type": "vsphere-iso",
				"network_adapters": []any{
					map[string]any{
						"network":      "",
						"network_card": "vmxnet3",
					},
				},
				"storage": []any{
					map[string]any{
						"disk_size": 5000,
					},
				},
			},
		},
		{
			Input: map[string]any{
				"type":                  "vsphere-iso",
				"network":               "myNetwork",
				"networkCard":           "vmxnet3",
				"disk_size":             5000,
				"disk_thin_provisioned": true,
				"disk_eagerly_scrub":    true,
			},

			Expected: map[string]any{
				"type": "vsphere-iso",
				"network_adapters": []any{
					map[string]any{
						"network":      "myNetwork",
						"network_card": "vmxnet3",
					},
				},
				"storage": []any{
					map[string]any{
						"disk_size":             5000,
						"disk_thin_provisioned": true,
						"disk_eagerly_scrub":    true,
					},
				},
			},
		},
		{
			Input: map[string]any{
				"type":                  "vsphere-iso",
				"network":               "myNetwork",
				"networkCard":           "vmxnet3",
				"disk_size":             5000,
				"disk_thin_provisioned": true,
				"disk_eagerly_scrub":    true,
				"network_adapters": []any{
					map[string]any{
						"network":      "net1",
						"network_card": "vmxnet3",
					},
				},
				"storage": []any{
					map[string]any{
						"disk_size":             5001,
						"disk_thin_provisioned": true,
						"disk_eagerly_scrub":    true,
					},
				},
			},

			Expected: map[string]any{
				"type": "vsphere-iso",
				"network_adapters": []any{
					map[string]any{
						"network":      "myNetwork",
						"network_card": "vmxnet3",
					},
					map[string]any{
						"network":      "net1",
						"network_card": "vmxnet3",
					},
				},
				"storage": []any{
					map[string]any{
						"disk_size":             5000,
						"disk_thin_provisioned": true,
						"disk_eagerly_scrub":    true,
					},
					map[string]any{
						"disk_size":             5001,
						"disk_thin_provisioned": true,
						"disk_eagerly_scrub":    true,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerVSphereNetworkDisk

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

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
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type":        "vsphere-iso",
				"network":     "",
				"networkCard": "vmxnet3",
				"disk_size":   5000,
			},

			Expected: map[string]interface{}{
				"type": "vsphere-iso",
				"network_adapters": []interface{}{
					map[string]interface{}{
						"network":     "",
						"networkCard": "vmxnet3",
					},
				},
				"storage": []interface{}{
					map[string]interface{}{
						"disk_size": 5000,
					},
				},
			},
		},
		{
			Input: map[string]interface{}{
				"type":        "vsphere-iso",
				"network":     "myNetwork",
				"networkCard": "vmxnet3",
				"disk_size":   5000,
			},

			Expected: map[string]interface{}{
				"type": "vsphere-iso",
				"network_adapters": []interface{}{
					map[string]interface{}{
						"network":     "myNetwork",
						"networkCard": "vmxnet3",
					},
				},
				"storage": []interface{}{
					map[string]interface{}{
						"disk_size": 5000,
					},
				},
			},
		},
		{
			Input: map[string]interface{}{
				"type":        "vsphere-iso",
				"network":     "myNetwork",
				"networkCard": "vmxnet3",
				"disk_size":   5000,
				"network_adapters": []interface{}{
					map[string]interface{}{
						"network":     "net1",
						"networkCard": "vmxnet3",
					},
				},
				"storage": []interface{}{
					map[string]interface{}{
						"disk_size": 5001,
					},
				},
			},

			Expected: map[string]interface{}{
				"type": "vsphere-iso",
				"network_adapters": []interface{}{
					map[string]interface{}{
						"network":     "myNetwork",
						"networkCard": "vmxnet3",
					},
					map[string]interface{}{
						"network":     "net1",
						"networkCard": "vmxnet3",
					},
				},
				"storage": []interface{}{
					map[string]interface{}{
						"disk_size": 5000,
					},
					map[string]interface{}{
						"disk_size": 5001,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerVSphereNetworkDisk

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

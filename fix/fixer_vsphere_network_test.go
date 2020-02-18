package fix

import (
	"reflect"
	"testing"
)

func TestFixerVSphereNetwork_impl(t *testing.T) {
	var _ Fixer = new(FixerVSphereNetwork)
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
			},

			Expected: map[string]interface{}{
				"type": "vsphere-iso",
				"network_adapters": []interface{}{map[string]interface{}{
					"network":     "",
					"networkCard": "vmxnet3",
				},
				},
			},
		},
		{
			Input: map[string]interface{}{
				"type":        "vsphere-iso",
				"network":     "myNetwork",
				"networkCard": "vmxnet3",
			},

			Expected: map[string]interface{}{
				"type": "vsphere-iso",
				"network_adapters": []interface{}{map[string]interface{}{
					"network":     "myNetwork",
					"networkCard": "vmxnet3",
				},
				},
			},
		},
		{
			Input: map[string]interface{}{
				"type":        "vsphere-iso",
				"network":     "myNetwork",
				"networkCard": "vmxnet3",
				"network_adapters": []interface{}{map[string]interface{}{
					"network":     "net1",
					"networkCard": "vmxnet3",
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
			},
		},
	}

	for _, tc := range cases {
		var f FixerVSphereNetwork

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

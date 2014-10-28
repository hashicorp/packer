package fix

import (
	"reflect"
	"testing"
)

func TestFixerVirtualBoxGAAttach_Impl(t *testing.T) {
	var _ Fixer = new(FixerVirtualBoxGAAttach)
}

func TestFixerVirtualBoxGAAttach_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No attach field
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox",
			},
		},

		// Attach field == false
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
				"guest_additions_attach": false,
			},

			Expected: map[string]interface{}{
				"type":                 "virtualbox",
				"guest_additions_mode": "upload",
			},
		},

		// Attach field == true
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
				"guest_additions_attach": true,
			},

			Expected: map[string]interface{}{
				"type":                 "virtualbox",
				"guest_additions_mode": "attach",
			},
		},

		// Attach field is not a bool
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
				"guest_additions_attach": "what",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox",
				"guest_additions_attach": "what",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxGAAttach

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

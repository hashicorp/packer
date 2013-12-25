package fix

import (
	"reflect"
	"testing"
)

func TestFixerVMwareRename_impl(t *testing.T) {
	var _ Fixer = new(FixerVMwareRename)
}

func TestFixerVMwareRename_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type": "vmware",
			},

			Expected: map[string]interface{}{
				"type": "vmware-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVMwareRename

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

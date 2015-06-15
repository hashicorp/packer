package fix

import (
	"reflect"
	"testing"
)

func TestFixerVirtualBoxRename_impl(t *testing.T) {
	var _ Fixer = new(FixerVirtualBoxRename)
}

func TestFixerVirtualBoxRename_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxRename

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

func TestFixerVirtualBoxRenameFix_provisionerOverride(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Input: map[string]interface{}{
				"provisioners": []interface{}{
					map[string]interface{}{
						"override": map[string]interface{}{
							"virtualbox": map[string]interface{}{},
						},
					},
				},
			},

			Expected: map[string]interface{}{
				"provisioners": []interface{}{
					map[string]interface{}{
						"override": map[string]interface{}{
							"virtualbox-iso": map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		var f FixerVirtualBoxRename

		output, err := f.Fix(tc.Input)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(output, tc.Expected) {
			t.Fatalf("unexpected:\n\n%#v\nexpected:\n\n%#v\n", output, tc.Expected)
		}
	}
}

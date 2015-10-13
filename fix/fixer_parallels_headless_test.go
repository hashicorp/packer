package fix

import (
	"reflect"
	"testing"
)

func TestFixerParallelsHeadless_Impl(t *testing.T) {
	var _ Fixer = new(FixerParallelsHeadless)
}

func TestFixerParallelsHeadless_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No headless field
		{
			Input: map[string]interface{}{
				"type": "parallels-iso",
			},

			Expected: map[string]interface{}{
				"type": "parallels-iso",
			},
		},

		// Headless field
		{
			Input: map[string]interface{}{
				"type":     "parallels-iso",
				"headless": false,
			},

			Expected: map[string]interface{}{
				"type": "parallels-iso",
			},
		},
	}

	for _, tc := range cases {
		var f FixerParallelsHeadless

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

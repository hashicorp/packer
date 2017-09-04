package fix

import (
	"reflect"
	"testing"
)

func TestFixerEnhancedNetworking_Impl(t *testing.T) {
	var _ Fixer = new(FixerEnhancedNetworking)
}

func TestFixerEnhancedNetworking(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// Attach field == false
		{
			Input: map[string]interface{}{
				"type":                "ebs",
				"enhanced_networking": false,
			},

			Expected: map[string]interface{}{
				"type":        "ebs",
				"ena_support": false,
			},
		},

		// Attach field == true
		{
			Input: map[string]interface{}{
				"type":                "ebs",
				"enhanced_networking": true,
			},

			Expected: map[string]interface{}{
				"type":        "ebs",
				"ena_support": true,
			},
		},
	}

	for _, tc := range cases {
		var f FixerEnhancedNetworking

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

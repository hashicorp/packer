package fix

import (
	"reflect"
	"testing"
)

func TestFixerCreateTime_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerCreateTime)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerCreateTime_Fix(t *testing.T) {
	var f FixerCreateTime

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]string{
				"type":     "foo",
				"ami_name": "{{.CreateTime}} foo",
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
			map[string]interface{}{
				"type":     "foo",
				"ami_name": "{{timestamp}} foo",
			},
		},
	}

	output, err := f.Fix(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
	}
}

package common

import (
	"reflect"
	"testing"
)

func TestTraverseStrings_InterfaceSlice(t *testing.T) {
	input := []interface{}{"foo", "bar"}

	f := func(string, string) string {
		return "bar"
	}

	err := TraverseStrings(&input, f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []interface{}{"bar", "bar"}
	if !reflect.DeepEqual(input, expected) {
		t.Fatalf("no: %#v", input)
	}
}

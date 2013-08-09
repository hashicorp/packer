package command

import (
	"flag"
	"reflect"
	"testing"
)

func TestSliceValue_implements(t *testing.T) {
	var raw interface{}
	raw = new(SliceValue)
	if _, ok := raw.(flag.Value); !ok {
		t.Fatalf("SliceValue should be a Value")
	}
}

func TestSliceValueSet(t *testing.T) {
	sv := new(SliceValue)
	err := sv.Set("foo,bar,baz")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual([]string(*sv), expected) {
		t.Fatalf("Bad: %#v", sv)
	}
}

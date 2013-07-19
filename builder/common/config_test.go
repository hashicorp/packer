package common

import (
	"reflect"
	"testing"
)

func TestDecodeConfig(t *testing.T) {
	type Local struct {
		Foo string
		Bar string
	}

	raws := []interface{}{
		map[string]interface{}{
			"foo": "bar",
		},
		map[string]interface{}{
			"bar": "baz",
			"baz": "what",
		},
	}

	var result Local
	md, err := DecodeConfig(&result, raws...)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Foo != "bar" {
		t.Fatalf("invalid: %#v", result.Foo)
	}

	if result.Bar != "baz" {
		t.Fatalf("invalid: %#v", result.Bar)
	}

	if md == nil {
		t.Fatal("metadata should not be nil")
	}

	if !reflect.DeepEqual(md.Unused, []string{"baz"}) {
		t.Fatalf("unused: %#v", md.Unused)
	}
}

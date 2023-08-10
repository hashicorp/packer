// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sliceflag

import (
	"flag"
	"reflect"
	"testing"
)

func TestStringFlag_implements(t *testing.T) {
	var raw interface{}
	raw = new(StringFlag)
	if _, ok := raw.(flag.Value); !ok {
		t.Fatalf("StringFlag should be a Value")
	}
}

// TestStringFlagSet tests for setting the same flag more than once on the CLI
// like: blah -flag foo -flag bar
func TestStringFlagSet(t *testing.T) {
	sv := new(StringFlag)
	err := sv.Set("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = sv.Set("bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual([]string(*sv), expected) {
		t.Fatalf("Bad: %#v", sv)
	}
}

// TestMultiStringFlag tests for setting the same flag using a comma-separated
// list of items like: blah -flag=foo,bar
func TestMultiStringFlag(t *testing.T) {
	sv := new(StringFlag)
	err := sv.Set("chocolate,vanilla")
	if err != nil {
		t.Fatalf("err :%s", err)
	}

	expected := []string{"chocolate", "vanilla"}
	if !reflect.DeepEqual([]string(*sv), expected) {
		t.Fatalf("Expected: %#v, found: %#v", expected, sv)
	}
}

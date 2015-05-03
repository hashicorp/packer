package common

import (
	"reflect"
	"testing"
)

func TestPrlctlPostConfigPrepare_PrlctlPost(t *testing.T) {
	// Test with empty
	c := new(PrlctlPostConfig)
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !reflect.DeepEqual(c.PrlctlPost, [][]string{}) {
		t.Fatalf("bad: %#v", c.PrlctlPost)
	}

	// Test with a good one
	c = new(PrlctlPostConfig)
	c.PrlctlPost = [][]string{
		{"foo", "bar", "baz"},
	}
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	expected := [][]string{
		[]string{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(c.PrlctlPost, expected) {
		t.Fatalf("bad: %#v", c.PrlctlPost)
	}
}

package common

import (
	"reflect"
	"testing"
)

func TestVBoxManageConfigPrepare_VBoxManage(t *testing.T) {
	// Test with empty
	c := new(VBoxManageConfig)
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !reflect.DeepEqual(c.VBoxManage, [][]string{}) {
		t.Fatalf("bad: %#v", c.VBoxManage)
	}

	// Test with a good one
	c = new(VBoxManageConfig)
	c.VBoxManage = [][]string{
		{"foo", "bar", "baz"},
	}
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	expected := [][]string{
		{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(c.VBoxManage, expected) {
		t.Fatalf("bad: %#v", c.VBoxManage)
	}
}

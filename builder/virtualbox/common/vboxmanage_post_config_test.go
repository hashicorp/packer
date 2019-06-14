package common

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/template/interpolate"
)

func TestVBoxManagePostConfigPrepare_VBoxManage(t *testing.T) {
	// Test with empty
	c := new(VBoxManagePostConfig)
	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !reflect.DeepEqual(c.VBoxManagePost, [][]string{}) {
		t.Fatalf("bad: %#v", c.VBoxManagePost)
	}

	// Test with a good one
	c = new(VBoxManagePostConfig)
	c.VBoxManagePost = [][]string{
		{"foo", "bar", "baz"},
	}
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	expected := [][]string{
		{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(c.VBoxManagePost, expected) {
		t.Fatalf("bad: %#v", c.VBoxManagePost)
	}
}

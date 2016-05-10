package common

import (
	"testing"
)

func TestVMXConfigPrepare(t *testing.T) {
	c := new(VMXConfig)
	c.VMXData = map[string]string{
		"one": "foo",
		"two": "bar",
	}

	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	if len(c.VMXData) != 2 {
		t.Fatal("should have two items in VMXData")
	}
}

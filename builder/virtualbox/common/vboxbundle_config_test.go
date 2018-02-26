package common

import (
	"reflect"
	"testing"
)

func TestVBoxBundleConfigPrepare_VBoxBundle(t *testing.T) {
	// Test with empty
	c := new(VBoxBundleConfig)
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !reflect.DeepEqual(c, VBoxBundleConfig{BundleISO: false}) {
		t.Fatalf("bad: %#v", c.VBoxBundle)
	}

	// Test with a good one
	c = new(VBoxBundleConfig)
	c.BundleISO = true
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	expected := VBoxBundleConfig{
    BundleISO: true,
	}

	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("bad: %#v", c)
	}
}

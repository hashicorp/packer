package common

import (
	"testing"
)

func TestVBoxVersionConfigPrepare_BootWait(t *testing.T) {
	var c *VBoxVersionConfig
	var errs []error

	// Test empty
	c = new(VBoxVersionConfig)
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.VBoxVersionFile != ".vbox_version" {
		t.Fatalf("bad value: %s", c.VBoxVersionFile)
	}

	// Test with a good one
	c = new(VBoxVersionConfig)
	c.VBoxVersionFile = "foo"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.VBoxVersionFile != "foo" {
		t.Fatalf("bad value: %s", c.VBoxVersionFile)
	}
}

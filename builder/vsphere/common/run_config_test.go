package common

import (
	"testing"
)

func TestRunConfigPrepare_BootWait(t *testing.T) {
	var c *RunConfig

	// Test a default boot_wait
	c = new(RunConfig)
	c.RawBootWait = ""
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
	if c.RawBootWait != "10s" {
		t.Fatalf("bad value: %s", c.RawBootWait)
	}

	// Test with a bad boot_wait
	c = new(RunConfig)
	c.RawBootWait = "this is not good"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatal("should error")
	}

	// Test with a good one
	c = new(RunConfig)
	c.RawBootWait = "5s"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

func TestRunConfigPrepare_VNCPort(t *testing.T) {
	var c *RunConfig

	// Bad
	c = new(RunConfig)
	c.VNCPortMin = 1000
	c.VNCPortMax = 500
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatal("should error")
	}

	// Bad
	c = new(RunConfig)
	c.VNCPortMin = 500
	c.VNCPortMax = 1000
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) != 0 {
		t.Fatal("should not error")
	}
}

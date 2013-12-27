package common

import (
	"testing"
)

func TestRunConfigPrepare(t *testing.T) {
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

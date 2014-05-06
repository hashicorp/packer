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

func TestRunConfigPrepare_CommunicatorType(t *testing.T) {
	var c *RunConfig

	// Test a default communicator_type
	c = new(RunConfig)
	c.CommunicatorType = ""
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
	if c.CommunicatorType != "ssh" {
		t.Fatalf("bad default communicator type: %s", c.CommunicatorType)
	}

	// Test with a bad communicator type
	c = new(RunConfig)
	c.CommunicatorType = "foo"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatal("should error")
	}

	// Test with ssh
	c = new(RunConfig)
	c.CommunicatorType = "ssh"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Test with winrm
	c = new(RunConfig)
	c.CommunicatorType = "winrm"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

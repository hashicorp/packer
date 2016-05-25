package common

import (
	"testing"
)

func TestRunConfigPrepare_BootWait(t *testing.T) {
	var c *RunConfig
	var errs []error

	// Test a default boot_wait
	c = new(RunConfig)
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.RawBootWait != "10s" {
		t.Fatalf("bad value: %s", c.RawBootWait)
	}

	// Test with a bad boot_wait
	c = new(RunConfig)
	c.RawBootWait = "this is not good"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Test with a good one
	c = new(RunConfig)
	c.RawBootWait = "5s"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

func TestRunConfigPrepare_VRDPBindAddress(t *testing.T) {
	var c *RunConfig
	var errs []error

	// Test a default VRDPBindAddress
	c = new(RunConfig)
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.VRDPBindAddress != "127.0.0.1" {
		t.Fatalf("bad value: %s", c.VRDPBindAddress)
	}

	// Test with a good one
	c = new(RunConfig)
	c.VRDPBindAddress = "192.168.0.1"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

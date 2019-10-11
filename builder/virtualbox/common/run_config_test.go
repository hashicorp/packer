package common

import (
	"testing"

	"github.com/hashicorp/packer/template/interpolate"
)

func TestRunConfigPrepare_VRDPBindAddress(t *testing.T) {
	var c *RunConfig
	var errs []error

	// Test a default VRDPBindAddress
	c = new(RunConfig)
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.VRDPBindAddress != "127.0.0.1" {
		t.Fatalf("bad value: %s", c.VRDPBindAddress)
	}

	// Test with a good one
	c = new(RunConfig)
	c.VRDPBindAddress = "192.168.0.1"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

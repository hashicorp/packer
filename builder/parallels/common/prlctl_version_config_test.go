package common

import (
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

func TestPrlctlVersionConfigPrepare_BootWait(t *testing.T) {
	var c *PrlctlVersionConfig
	var errs []error

	// Test empty
	c = new(PrlctlVersionConfig)
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.PrlctlVersionFile != ".prlctl_version" {
		t.Fatalf("bad value: %s", c.PrlctlVersionFile)
	}

	// Test with a good one
	c = new(PrlctlVersionConfig)
	c.PrlctlVersionFile = "foo"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.PrlctlVersionFile != "foo" {
		t.Fatalf("bad value: %s", c.PrlctlVersionFile)
	}
}

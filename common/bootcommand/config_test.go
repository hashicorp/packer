package bootcommand

import (
	"testing"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

func TestConfigPrepare(t *testing.T) {
	var c *BootConfig

	// Test a default boot_wait
	c = new(BootConfig)
	c.BootWait = 0
	errs := c.Prepare(&interpolate.Context{})
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
	if c.BootWait != 10*time.Second {
		t.Fatalf("bad value: %s", c.BootWait)
	}

	// Test with a good one
	c = new(BootConfig)
	c.BootWait = 5 * time.Second
	errs = c.Prepare(&interpolate.Context{})
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

func TestVNCConfigPrepare(t *testing.T) {
	var c *VNCConfig

	// Test with a boot command
	c = new(VNCConfig)
	c.BootCommand = []string{"a", "b"}
	errs := c.Prepare(&interpolate.Context{})
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Test with disabled vnc
	c.DisableVNC = true
	errs = c.Prepare(&interpolate.Context{})
	if len(errs) == 0 {
		t.Fatal("should error")
	}

	// Test no boot command with no vnc
	c = new(VNCConfig)
	c.DisableVNC = true
	errs = c.Prepare(&interpolate.Context{})
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

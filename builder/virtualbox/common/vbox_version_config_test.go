package common

import (
	"testing"
)

func TestVBoxVersionConfigPrepare_BootWait(t *testing.T) {
	var c *VBoxVersionConfig
	var errs []error

	// Test empty
	c = new(VBoxVersionConfig)
	errs = c.Prepare("ssh")
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if *c.VBoxVersionFile != ".vbox_version" {
		t.Fatalf("bad value: %s", *c.VBoxVersionFile)
	}

	// Test with a good one
	c = new(VBoxVersionConfig)
	filename := "foo"
	c.VBoxVersionFile = &filename
	errs = c.Prepare("ssh")
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if *c.VBoxVersionFile != "foo" {
		t.Fatalf("bad value: %s", *c.VBoxVersionFile)
	}
}

func TestVBoxVersionConfigPrepare_empty(t *testing.T) {
	var c *VBoxVersionConfig
	var errs []error

	// Test with nil value
	c = new(VBoxVersionConfig)
	c.VBoxVersionFile = nil
	errs = c.Prepare("ssh")
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if *c.VBoxVersionFile != ".vbox_version" {
		t.Fatalf("bad value: %s", *c.VBoxVersionFile)
	}

	// Test with empty name
	c = new(VBoxVersionConfig)
	filename := ""
	c.VBoxVersionFile = &filename
	errs = c.Prepare("ssh")
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if *c.VBoxVersionFile != "" {
		t.Fatalf("bad value: %s", *c.VBoxVersionFile)
	}
}

func TestVBoxVersionConfigPrepare_communicator(t *testing.T) {
	var c *VBoxVersionConfig
	var errs []error

	// Test with 'none' communicator and non-empty virtualbox_version_file
	c = new(VBoxVersionConfig)
	filename := "test"
	c.VBoxVersionFile = &filename
	errs = c.Prepare("none")
	if len(errs) == 0 {
		t.Fatalf("should have an error")
	}
}

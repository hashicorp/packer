package common

import (
	"testing"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

func testShutdownConfig() *ShutdownConfig {
	return &ShutdownConfig{}
}

func TestShutdownConfigPrepare_ShutdownCommand(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	c = testShutdownConfig()
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
}

func TestShutdownConfigPrepare_ShutdownTimeout(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	// Test with a good one
	c = testShutdownConfig()
	c.ShutdownTimeout = 5 * time.Second
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.ShutdownTimeout != 5*time.Second {
		t.Fatalf("bad: %s", c.ShutdownTimeout)
	}
}

func TestShutdownConfigPrepare_PostShutdownDelay(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	// Test with default value
	c = testShutdownConfig()
	c.PostShutdownDelay = 0
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay.Nanoseconds() != 0 {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}

	// Test with a good one
	c = testShutdownConfig()
	c.PostShutdownDelay = 5 * time.Millisecond
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay != 5*time.Millisecond {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}
}

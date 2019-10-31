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

	// Test with a bad value
	c = testShutdownConfig()
	c.ShutdownTimeout = "this is not good"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	// Test with a good one
	c = testShutdownConfig()
	c.ShutdownTimeout = "5s"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.ShutdownTimeout.Duration() != 5*time.Second {
		t.Fatalf("bad: %s", c.ShutdownTimeout)
	}
}

func TestShutdownConfigPrepare_PostShutdownDelay(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	// Test with a bad value
	c = testShutdownConfig()
	c.PostShutdownDelay = "this is not good"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	// Test with default value
	c = testShutdownConfig()
	c.PostShutdownDelay = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay.Duration().Nanoseconds() != 0 {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}

	// Test with a good one
	c = testShutdownConfig()
	c.PostShutdownDelay = "5s"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay.Duration() != 5*time.Second {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}
}

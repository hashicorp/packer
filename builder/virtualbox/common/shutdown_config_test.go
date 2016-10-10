package common

import (
	"testing"
	"time"
)

func testShutdownConfig() *ShutdownConfig {
	return &ShutdownConfig{}
}

func TestShutdownConfigPrepare_ShutdownCommand(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	c = testShutdownConfig()
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
}

func TestShutdownConfigPrepare_ShutdownTimeout(t *testing.T) {
	var c *ShutdownConfig
	var errs []error

	// Test with a bad value
	c = testShutdownConfig()
	c.RawShutdownTimeout = "this is not good"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	// Test with a good one
	c = testShutdownConfig()
	c.RawShutdownTimeout = "5s"
	errs = c.Prepare(testConfigTemplate(t))
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

	// Test with a bad value
	c = testShutdownConfig()
	c.RawPostShutdownDelay = "this is not good"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	// Test with default value
	c = testShutdownConfig()
	c.RawPostShutdownDelay = ""
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay.Nanoseconds() != 0 {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}

	// Test with a good one
	c = testShutdownConfig()
	c.RawPostShutdownDelay = "5s"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.PostShutdownDelay != 5*time.Second {
		t.Fatalf("bad: %s", c.PostShutdownDelay)
	}
}

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
	var warns []string
	var errs []error

	c = testShutdownConfig()
	c.ShutdownCommand = "foo"
	warns, errs = c.Prepare(testConfigTemplate(t))
	if len(warns) > 0 {
		t.Fatalf("warn: %#v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
}

func TestShutdownConfigPrepare_NoShutdownCommand(t *testing.T) {
	var c *ShutdownConfig
	var warns []string
	var errs []error

	c = testShutdownConfig()
	c.ShutdownCommand = ""
	warns, errs = c.Prepare(testConfigTemplate(t))
	if len(warns) == 0 {
		t.Fatalf("Should warn")
	}
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
	_, errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	// Test with a good one
	c = testShutdownConfig()
	c.RawShutdownTimeout = "5s"
	_, errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if c.ShutdownTimeout != 5*time.Second {
		t.Fatalf("bad: %s", c.ShutdownTimeout)
	}
}

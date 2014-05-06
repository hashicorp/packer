package common

import (
	"testing"
)

func testWinRMConfig() *WinRMConfig {
	return &WinRMConfig{
		WinRMUser: "admin",
	}
}

func TestWinRMConfigPrepare(t *testing.T) {
	c := testWinRMConfig()
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.WinRMPort != 5985 {
		t.Errorf("bad winrm port: %d", c.WinRMPort)
	}
}

func TestWinRMConfigPrepare_WinRMUser(t *testing.T) {
	var c *WinRMConfig
	var errs []error

	c = testWinRMConfig()
	c.WinRMUser = ""
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("should have error")
	}

	c = testWinRMConfig()
	c.WinRMUser = "exists"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
}

func TestWinRMConfigPrepare_WinRMWaitTimeout(t *testing.T) {
	var c *WinRMConfig
	var errs []error

	// Defaults
	c = testWinRMConfig()
	c.RawWinRMWaitTimeout = ""
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if c.RawWinRMWaitTimeout != "20m" {
		t.Fatalf("bad value: %s", c.RawWinRMWaitTimeout)
	}

	// Test with a bad value
	c = testWinRMConfig()
	c.RawWinRMWaitTimeout = "this is not good"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatal("should have error")
	}

	// Test with a good one
	c = testWinRMConfig()
	c.RawWinRMWaitTimeout = "5s"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
}

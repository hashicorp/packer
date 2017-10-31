package common

import (
	"testing"
)

func TestExportConfigPrepareCorrect(t *testing.T) {
	var c *ExportConfig

	// Test with a good one
	c = new(ExportConfig)
	c.Format = "vmx"
	c.SkipExport = false
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

func TestExportConfigPrepareSkip(t *testing.T) {
	var c *ExportConfig

	// Test with a good one
	c = new(ExportConfig)
	c.SkipExport = true
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

func TestExportConfigPrepareDefault(t *testing.T) {
	var c *ExportConfig

	// Test with a good one
	c = new(ExportConfig)
	c.Format = ""
	c.SkipExport = false
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	if c.Format != "ovf" {
		t.Fatalf("Format by default should be ovf: %#v", c.Format)
	}
}
func TestExportConfigPrepareIncorrect(t *testing.T) {
	var c *ExportConfig

	c = new(ExportConfig)
	c.Format = "foo"
	c.SkipExport = false
	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) == 0 {
		t.Fatalf("export config should error")
	}
}

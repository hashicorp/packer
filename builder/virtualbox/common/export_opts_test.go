package common

import (
	"testing"
)

func TestExportOptsPrepare_BootWait(t *testing.T) {
	var c *ExportOpts
	var errs []error

	// Good
	c = new(ExportOpts)
	c.ExportOpts = []string{
		"--options",
	}
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

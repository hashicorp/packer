package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestExportConfigPrepare_BootWait(t *testing.T) {
	var c *ExportConfig
	var errs []error

	// Bad
	c = new(ExportConfig)
	c.Format = "illegal"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Good
	c = new(ExportConfig)
	c.Format = "ova"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	// Good
	c = new(ExportConfig)
	c.Format = "ovf"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

func TestExportConfigPrepare_Opts(t *testing.T) {
	var c *ExportConfig
	var errs []error

	// Good
	c = new(ExportConfig)
	c.ExportOpts = []string{
		"--options",
	}
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

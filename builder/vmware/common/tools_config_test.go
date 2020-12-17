package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestToolsConfigPrepare_Empty(t *testing.T) {
	c := &ToolsConfig{}

	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.ToolsUploadPath != "{{ .Flavor }}.iso" {
		t.Fatal("should have defaulted tools upload path")
	}
}

func TestToolsConfigPrepare_SetUploadPath(t *testing.T) {
	c := &ToolsConfig{
		ToolsUploadPath: "path/to/tools.iso",
	}

	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.ToolsUploadPath != "path/to/tools.iso" {
		t.Fatal("should have used given tools upload path")
	}
}

func TestToolsConfigPrepare_ErrorIfOnlySource(t *testing.T) {
	c := &ToolsConfig{
		ToolsSourcePath: "path/to/tools.iso",
	}

	errs := c.Prepare(interpolate.NewContext())
	if len(errs) != 1 {
		t.Fatalf("Should have received an error because the flavor and " +
			"upload path aren't set")
	}
}

func TestToolsConfigPrepare_SourceSuccess(t *testing.T) {
	for _, c := range []*ToolsConfig{
		&ToolsConfig{
			ToolsSourcePath: "path/to/tools.iso",
			ToolsUploadPath: "partypath.iso",
		},
		&ToolsConfig{
			ToolsSourcePath:   "path/to/tools.iso",
			ToolsUploadFlavor: "linux",
		},
	} {
		errs := c.Prepare(interpolate.NewContext())
		if len(errs) != 0 {
			t.Fatalf("Should not have received an error")
		}
	}
}

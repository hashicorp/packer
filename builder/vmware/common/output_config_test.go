package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestOutputConfigPrepare(t *testing.T) {
	c := new(OutputConfig)
	if c.OutputDir != "" {
		t.Fatalf("what: %s", c.OutputDir)
	}

	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(interpolate.NewContext(), pc)
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.OutputDir == "" {
		t.Fatal("should have output dir")
	}
}

package common

import (
	"testing"

	"github.com/hashicorp/packer/common"
)

func TestOutputConfigPrepare(t *testing.T) {
	c := new(OutputConfig)
	if c.OutputDir != "" {
		t.Fatalf("what: %s", c.OutputDir)
	}

	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.OutputDir == "" {
		t.Fatal("should have output dir")
	}
}

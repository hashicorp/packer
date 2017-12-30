package common

import (
	"github.com/hashicorp/packer/common"
	"io/ioutil"
	"os"
	"testing"
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

func TestOutputConfigPrepare_exists(t *testing.T) {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	c := new(OutputConfig)
	c.OutputDir = td

	pc := &common.PackerConfig{
		PackerBuildName: "foo",
		PackerForce:     false,
	}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) != 0 {
		t.Fatal("should not have errors")
	}
}

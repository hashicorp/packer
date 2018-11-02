// +build !windows

package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer/configfile"
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
	prefix, _ := configfile.ConfigTmpDir()
	td, err := ioutil.TempDir(prefix, "parallels")
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
	if len(errs) == 0 {
		t.Fatal("should have errors")
	}
}

func TestOutputConfigPrepare_forceExists(t *testing.T) {
	prefix, _ := configfile.ConfigTmpDir()
	td, err := ioutil.TempDir(prefix, "parallels")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	c := new(OutputConfig)
	c.OutputDir = td

	pc := &common.PackerConfig{
		PackerBuildName: "foo",
		PackerForce:     true,
	}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatal("should not have errors")
	}
}

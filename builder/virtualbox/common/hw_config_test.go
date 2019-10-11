package common

import (
	"testing"

	"github.com/hashicorp/packer/template/interpolate"
)

func TestHWConfigPrepare(t *testing.T) {
	c := new(HWConfig)
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.CpuCount < 1 {
		t.Errorf("bad cpu count: %d", c.CpuCount)
	}

	if c.MemorySize < 64 {
		t.Errorf("bad memory size: %d", c.MemorySize)
	}
}

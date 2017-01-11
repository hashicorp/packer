package common

import (
	"testing"
)

func TestFloppyConfigPrepare(t *testing.T) {
	c := new(FloppyConfig)

	errs := c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if len(c.FloppyFiles) > 0 {
		t.Fatal("should not have floppy files")
	}
}

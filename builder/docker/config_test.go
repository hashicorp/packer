package docker

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfigStruct(t *testing.T) *Config {
	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return &Config{
		ExportPath: "foo",
		Image:      "bar",
		tpl:        tpl,
	}
}

func TestConfigPrepare_exportPath(t *testing.T) {
	c := testConfigStruct(t)

	// No export path
	c.ExportPath = ""
	warns, errs := c.Prepare()
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if len(errs) <= 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Good export path
	c.ExportPath = "path"
	warns, errs = c.Prepare()
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

func TestConfigPrepare_image(t *testing.T) {
	c := testConfigStruct(t)

	// No image
	c.Image = ""
	warns, errs := c.Prepare()
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if len(errs) <= 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Good image
	c.Image = "path"
	warns, errs = c.Prepare()
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}
}

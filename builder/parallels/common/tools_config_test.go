package common

import (
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

func testToolsConfig() *ToolsConfig {
	return &ToolsConfig{
		ParallelsToolsFlavor:    "foo",
		ParallelsToolsGuestPath: "foo",
		ParallelsToolsMode:      "attach",
	}
}

func TestToolsConfigPrepare(t *testing.T) {
	c := testToolsConfig()
	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("bad err: %#v", errs)
	}
}

func TestToolsConfigPrepare_ParallelsToolsMode(t *testing.T) {
	var c *ToolsConfig
	var errs []error

	// Test default mode
	c = testToolsConfig()
	c.ParallelsToolsMode = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if c.ParallelsToolsMode != ParallelsToolsModeUpload {
		t.Errorf("bad parallels tools mode: %s", c.ParallelsToolsMode)
	}

	// Test another mode
	c = testToolsConfig()
	c.ParallelsToolsMode = "attach"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if c.ParallelsToolsMode != ParallelsToolsModeAttach {
		t.Fatalf("bad mode: %s", c.ParallelsToolsMode)
	}

	// Test invalid mode
	c = testToolsConfig()
	c.ParallelsToolsMode = "invalid_mode"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}
}

func TestToolsConfigPrepare_ParallelsToolsGuestPath(t *testing.T) {
	var c *ToolsConfig
	var errs []error

	// Test default path
	c = testToolsConfig()
	c.ParallelsToolsGuestPath = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if c.ParallelsToolsGuestPath == "" {
		t.Fatal("should not be empty")
	}

	// Test with a good one
	c = testToolsConfig()
	c.ParallelsToolsGuestPath = "foo"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}

	if c.ParallelsToolsGuestPath != "foo" {
		t.Fatalf("bad guest path: %s", c.ParallelsToolsGuestPath)
	}
}

func TestToolsConfigPrepare_ParallelsToolsFlavor(t *testing.T) {
	var c *ToolsConfig
	var errs []error

	// Test with a default value
	c = testToolsConfig()
	c.ParallelsToolsFlavor = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}

	// Test with an bad value
	c = testToolsConfig()
	c.ParallelsToolsMode = "attach"
	c.ParallelsToolsFlavor = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}

	// Test with a good one
	c = testToolsConfig()
	c.ParallelsToolsMode = "disable"
	c.ParallelsToolsFlavor = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

package openstack

import (
	"testing"
)

func testImageConfig() *ImageConfig {
	return &ImageConfig{
		ImageName: "foo",
	}
}

func TestImageConfigPrepare_Region(t *testing.T) {
	c := testImageConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ImageName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

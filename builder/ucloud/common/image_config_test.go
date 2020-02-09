package common

import (
	"testing"
)

func testImageConfig() *ImageConfig {
	return &ImageConfig{
		ImageName: "foo",
	}
}

func TestImageConfigPrepare_name(t *testing.T) {
	c := testImageConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ImageName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestImageConfigPrepare_destinations(t *testing.T) {
	c := testImageConfig()
	c.ImageDestinations = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ImageDestinations = []ImageDestination{
		{
			ProjectId: "foo",
			Region:    "cn-bj2",
			Name:      "bar",
		},

		{
			ProjectId: "bar",
			Region:    "cn-sh2",
			Name:      "foo",
		},
	}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	c.ImageDestinations = []ImageDestination{
		{
			ProjectId: "foo",
			Name:      "bar",
		},

		{
			ProjectId: "bar",
			Region:    "cn-sh2",
		},
	}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

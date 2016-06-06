package common

import (
	"testing"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := testAccessConfig()
	c.RawRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.RawRegion = "us-east-12"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.RawRegion = "us-east-1"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.RawRegion = "custom"
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.RawRegion = "custom"
	c.SkipValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
	c.SkipValidation = false

}

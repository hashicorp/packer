package common

import (
	"testing"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := testAccessConfig()

	c.RawRegion = "us-east-12"
	err := c.ValidateOSCRegion(c.RawRegion)
	if err == nil {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "us-east-1"
	err = c.ValidateOSCRegion(c.RawRegion)
	if err == nil {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "custom"
	err = c.ValidateOSCRegion(c.RawRegion)
	if err == nil {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "custom"
	c.SkipValidation = true
	// testing whole prepare func here; this is checking that validation is
	// skipped, so we don't need a mock connection
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.SkipValidation = false
	c.RawRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

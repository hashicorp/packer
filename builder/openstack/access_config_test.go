package openstack

import (
	"testing"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_NoRegion(t *testing.T) {
	c := testAccessConfig()
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	dfw := "DFW"
	c := testAccessConfig()
	c.RawRegion = dfw
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
	if dfw != c.Region() {
		t.Fatalf("Regions do not match: %s %s", dfw, c.Region())
	}
}

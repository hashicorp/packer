package openstack

import (
	"testing"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := testAccessConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

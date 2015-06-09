package openstack

import (
	"os"
	"testing"
)

func init() {
	// Clear out the openstack env vars so they don't
	// affect our tests.
	os.Setenv("SDK_REGION", "")
	os.Setenv("OS_REGION_NAME", "")
}

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_NoRegion_Rackspace(t *testing.T) {
	c := testAccessConfig()
	c.Provider = "rackspace-us"
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

func TestAccessConfigRegionWithEmptyEnv(t *testing.T) {
	c := testAccessConfig()
	c.Prepare(nil)
	if c.Region() != "" {
		t.Fatalf("Region should be empty")
	}
}

func TestAccessConfigRegionWithSdkRegionEnv(t *testing.T) {
	c := testAccessConfig()
	c.Prepare(nil)

	expectedRegion := "sdk_region"
	os.Setenv("SDK_REGION", expectedRegion)
	os.Setenv("OS_REGION_NAME", "")
	if c.Region() != expectedRegion {
		t.Fatalf("Region should be: %s", expectedRegion)
	}
}

func TestAccessConfigRegionWithOsRegionNameEnv(t *testing.T) {
	c := testAccessConfig()
	c.Prepare(nil)

	expectedRegion := "os_region_name"
	os.Setenv("SDK_REGION", "")
	os.Setenv("OS_REGION_NAME", expectedRegion)
	if c.Region() != expectedRegion {
		t.Fatalf("Region should be: %s", expectedRegion)
	}
}

func TestAccessConfigPrepare_NoRegion_PrivateCloud(t *testing.T) {
	c := testAccessConfig()
	c.Provider = "http://some-keystone-server:5000/v2.0"
	if err := c.Prepare(nil); err != nil {
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

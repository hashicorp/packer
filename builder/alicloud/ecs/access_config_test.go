package ecs

import (
	"os"
	"testing"
)

func testAlicloudAccessConfig() *AlicloudAccessConfig {
	return &AlicloudAccessConfig{
		AlicloudAccessKey: "ak",
		AlicloudSecretKey: "acs",
	}

}

func TestAlicloudAccessConfigPrepareRegion(t *testing.T) {
	c := testAlicloudAccessConfig()

	c.AlicloudRegion = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.AlicloudRegion = "cn-beijing"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	os.Setenv("ALICLOUD_REGION", "cn-hangzhou")
	c.AlicloudRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudSkipValidation = false
}

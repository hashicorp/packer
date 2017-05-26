package ecs

import (
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
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudRegion = "cn-beijing-3"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.AlicloudRegion = "cn-beijing"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudRegion = "unknown"
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.AlicloudRegion = "unknown"
	c.AlicloudSkipValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
	c.AlicloudSkipValidation = false

}

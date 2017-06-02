package ecs

import (
	"testing"

	"github.com/denverdino/aliyungo/common"
)

func testAlicloudImageConfig() *AlicloudImageConfig {
	return &AlicloudImageConfig{
		AlicloudImageName: "foo",
	}
}

func TestECSImageConfigPrepare_name(t *testing.T) {
	c := testAlicloudImageConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudImageName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestAMIConfigPrepare_regions(t *testing.T) {
	c := testAlicloudImageConfig()
	c.AlicloudImageDestinationRegions = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudImageDestinationRegions = regionsToString()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudImageDestinationRegions = []string{"foo"}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.AlicloudImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou", "eu-central-1"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	c.AlicloudImageDestinationRegions = []string{"unknow"}
	c.AlicloudImageSkipRegionValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatal("shouldn't have error")
	}
	c.AlicloudImageSkipRegionValidation = false

}

func regionsToString() []string {
	var regions []string
	for _, region := range common.ValidRegions {
		regions = append(regions, string(region))
	}
	return regions
}

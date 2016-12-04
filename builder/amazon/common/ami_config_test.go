package common

import (
	"reflect"
	"testing"
)

func testAMIConfig() *AMIConfig {
	return &AMIConfig{
		AMIName: "foo",
	}
}

func TestAMIConfigPrepare_name(t *testing.T) {
	c := testAMIConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AMIName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestAMIConfigPrepare_regions(t *testing.T) {
	c := testAMIConfig()
	c.AMIRegions = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AMIRegions = listEC2Regions()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AMIRegions = []string{"foo"}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.AMIRegions = []string{"us-east-1", "us-west-1", "us-east-1"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	expected := []string{"us-east-1", "us-west-1"}
	if !reflect.DeepEqual(c.AMIRegions, expected) {
		t.Fatalf("bad: %#v", c.AMIRegions)
	}

	c.AMIRegions = []string{"custom"}
	c.AMISkipRegionValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatal("shouldn't have error")
	}
	c.AMISkipRegionValidation = false

}

func TestAMIConfigPrepare_Share_EncryptedBoot(t *testing.T) {
	c := testAMIConfig()
	c.AMIUsers = []string{"testAccountID"}
	c.AMIEncryptBootVolume = true

	c.AMIKmsKeyId = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("shouldn't be able to share ami with encrypted boot volume")
	}

	c.AMIKmsKeyId = "89c3fb9a-de87-4f2a-aedc-fddc5138193c"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("shouldn't be able to share ami with encrypted boot volume")
	}
}

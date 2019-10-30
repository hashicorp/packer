package common

import (
	"os"
	"testing"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{
		PublicKey:  "test_pub",
		PrivateKey: "test_pri",
		ProjectId:  "test_pro",
	}

}

func TestAccessConfigPrepareRegion(t *testing.T) {
	c := testAccessConfig()

	c.Region = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.Region = "cn-sh2"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	os.Setenv("UCLOUD_REGION", "cn-bj2")
	c.Region = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

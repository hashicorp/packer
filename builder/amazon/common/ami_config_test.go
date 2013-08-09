package common

import (
	"testing"
)

func testAMIConfig() *AMIConfig {
	return &AMIConfig{
		AMIName: "foo",
	}
}

func TestAMIConfigPrepare_Region(t *testing.T) {
	c := testAMIConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AMIName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

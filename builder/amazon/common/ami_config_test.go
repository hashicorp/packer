package common

import (
	"reflect"
	"testing"
	"time"
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
}

func TestAMIConfigPrepare_amiCopyTimeout(t *testing.T) {
	c := testAMIConfig()
	c.RawAMICopyTimeout = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.RawAMICopyTimeout = "30mm"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.RawAMICopyTimeout = "30m"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	expected, _ := time.ParseDuration("30m")
	if !reflect.DeepEqual(c.AMICopyTimeout(), expected) {
		t.Fatalf("bad: %#v", c.AMICopyTimeout())
	}
}

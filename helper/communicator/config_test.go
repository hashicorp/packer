package communicator

import (
	"testing"

	"github.com/mitchellh/packer/template/interpolate"
)

func testConfig() *Config {
	return &Config{
		SSHUsername: "root",
	}
}

func TestConfigType(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.Type != "ssh" {
		t.Fatal("bad: %#v", c)
	}
}

func TestConfig_none(t *testing.T) {
	c := &Config{Type: "none"}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}
}

func testContext(t *testing.T) *interpolate.Context {
	return nil
}

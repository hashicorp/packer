package openstack

import (
	"os"
	"testing"
)

func init() {
	// Clear out the openstack env vars so they don't
	// affect our tests.
	os.Setenv("SDK_USERNAME", "")
	os.Setenv("SDK_PASSWORD", "")
	os.Setenv("SDK_PROVIDER", "")
}

func testRunConfig() *RunConfig {
	return &RunConfig{
		SourceImage: "abcd",
		Flavor:      "m1.small",
		SSHUsername: "root",
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := testRunConfig()
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_InstanceType(t *testing.T) {
	c := testRunConfig()
	c.Flavor = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SourceImage(t *testing.T) {
	c := testRunConfig()
	c.SourceImage = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testRunConfig()
	c.SSHPort = 0
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.SSHPort)
	}

	c.SSHPort = 44
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.SSHPort)
	}
}

func TestRunConfigPrepare_SSHTimeout(t *testing.T) {
	c := testRunConfig()
	c.RawSSHTimeout = ""
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	c.RawSSHTimeout = "bad"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHUsername(t *testing.T) {
	c := testRunConfig()
	c.SSHUsername = ""
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

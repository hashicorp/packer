package common

import (
	"testing"
)

func testConfig() *RunConfig {
	return &RunConfig{
		Region: "us-east-1",
		SourceAmi: "abcd",
		InstanceType: "m1.small",
		SSHUsername: "root",
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := testConfig()
	err := c.Prepare()
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_InstanceType(t *testing.T) {
	c := testConfig()
	c.InstanceType = ""
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_Region(t *testing.T) {
	c := testConfig()
	c.Region = ""
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	c.Region = "us-east-12"
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	c.Region = "us-east-1"
	if err := c.Prepare(); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SourceAmi(t *testing.T) {
	c := testConfig()
	c.SourceAmi = ""
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testConfig()
	c.SSHPort = 0
	if err := c.Prepare(); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.SSHPort)
	}

	c.SSHPort = 44
	if err := c.Prepare(); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.SSHPort)
	}
}

func TestRunConfigPrepare_SSHTimeout(t *testing.T) {
	c := testConfig()
	c.RawSSHTimeout = ""
	if err := c.Prepare(); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	c.RawSSHTimeout = "bad"
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHUsername(t *testing.T) {
	c := testConfig()
	c.SSHUsername = ""
	if err := c.Prepare(); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

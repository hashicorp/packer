package common

import (
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	// Clear out the AWS access key env vars so they don't
	// affect our tests.
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_ACCESS_KEY", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	os.Setenv("AWS_SECRET_KEY", "")
}

func testConfig() *RunConfig {
	return &RunConfig{
		SourceAmi:    "abcd",
		InstanceType: "m1.small",
		SSHUsername:  "root",
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := testConfig()
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_InstanceType(t *testing.T) {
	c := testConfig()
	c.InstanceType = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SourceAmi(t *testing.T) {
	c := testConfig()
	c.SourceAmi = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testConfig()
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
	c := testConfig()
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
	c := testConfig()
	c.SSHUsername = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_UserData(t *testing.T) {
	c := testConfig()
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	c.UserData = "foo"
	c.UserDataFile = tf.Name()
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_UserDataFile(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	c.UserDataFile = "idontexistidontthink"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	c.UserDataFile = tf.Name()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_TemporaryKeyPairName(t *testing.T) {
	c := testConfig()
	c.TemporaryKeyPairName = ""
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.TemporaryKeyPairName == "" {
		t.Fatal("keypair empty")
	}
}

package ecs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/communicator"
)

func testConfig() *RunConfig {
	return &RunConfig{
		AlicloudSourceImage: "alicloud_images",
		InstanceType:        "ecs.n1.tiny",
		Comm: communicator.Config{
			SSHUsername: "alicloud",
		},
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

func TestRunConfigPrepare_SourceECSImage(t *testing.T) {
	c := testConfig()
	c.AlicloudSourceImage = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testConfig()
	c.Comm.SSHPort = 0
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}

	c.Comm.SSHPort = 44
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}
}

func TestRunConfigPrepare_UserData(t *testing.T) {
	c := testConfig()
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
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
	defer os.Remove(tf.Name())
	defer tf.Close()

	c.UserDataFile = tf.Name()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_TemporaryKeyPairName(t *testing.T) {
	c := testConfig()
	c.Comm.SSHTemporaryKeyPairName = ""
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHTemporaryKeyPairName == "" {
		t.Fatal("keypair name is empty")
	}

	c.Comm.SSHTemporaryKeyPairName = "ssh-key-123"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHTemporaryKeyPairName != "ssh-key-123" {
		t.Fatal("keypair name does not match")
	}
}

func TestRunConfigPrepare_SSHPrivateIp(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.SSHPrivateIp != false {
		t.Fatalf("invalid value, expected: %t, actul: %t", false, c.SSHPrivateIp)
	}
	c.SSHPrivateIp = true
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.SSHPrivateIp != true {
		t.Fatalf("invalid value, expected: %t, actul: %t", true, c.SSHPrivateIp)
	}
	c.SSHPrivateIp = false
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.SSHPrivateIp != false {
		t.Fatalf("invalid value, expected: %t, actul: %t", false, c.SSHPrivateIp)
	}
}

func TestRunConfigPrepare_DisableStopInstance(t *testing.T) {
	c := testConfig()

	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.DisableStopInstance != false {
		t.Fatalf("invalid value, expected: %t, actul: %t", false, c.DisableStopInstance)
	}

	c.DisableStopInstance = true
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.DisableStopInstance != true {
		t.Fatalf("invalid value, expected: %t, actul: %t", true, c.DisableStopInstance)
	}

	c.DisableStopInstance = false
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.DisableStopInstance != false {
		t.Fatalf("invalid value, expected: %t, actul: %t", false, c.DisableStopInstance)
	}
}

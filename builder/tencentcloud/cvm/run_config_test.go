package cvm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/communicator"
)

func testConfig() *TencentCloudRunConfig {
	return &TencentCloudRunConfig{
		SourceImageId: "img-qwer1234",
		InstanceType:  "S3.SMALL2",
		Comm: communicator.Config{
			SSHUsername: "tencentcloud",
		},
	}
}

func TestTencentCloudRunConfig_Prepare(t *testing.T) {
	cf := testConfig()

	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %v", err)
	}

	cf.InstanceType = ""
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should have err")
	}

	cf.InstanceType = "S3.SMALL2"
	cf.SourceImageId = ""
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should have err")
	}

	cf.SourceImageId = "img-qwer1234"
	cf.Comm.SSHPort = 0
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %v", err)
	}

	if cf.Comm.SSHPort != 22 {
		t.Fatalf("invalid ssh port value: %v", cf.Comm.SSHPort)
	}

	cf.Comm.SSHPort = 44
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %v", err)
	}

	if cf.Comm.SSHPort != 44 {
		t.Fatalf("invalid ssh port value: %v", cf.Comm.SSHPort)
	}
}

func TestTencentCloudRunConfigPrepare_UserData(t *testing.T) {
	cf := testConfig()
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("new temp file failed: %v", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	cf.UserData = "text user_data"
	cf.UserDataFile = tf.Name()
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestTencentCloudRunConfigPrepare_UserDataFile(t *testing.T) {
	cf := testConfig()
	cf.UserDataFile = "not-exist-file"
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("new temp file failed: %v", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	cf.UserDataFile = tf.Name()
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have error: %v", err)
	}
}

func TestTencentCloudRunConfigPrepare_TemporaryKeyPairName(t *testing.T) {
	cf := testConfig()
	cf.Comm.SSHTemporaryKeyPairName = ""
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have error: %v", err)
	}

	if cf.Comm.SSHTemporaryKeyPairName == "" {
		t.Fatal("invalid ssh key pair value")
	}

	cf.Comm.SSHTemporaryKeyPairName = "ssh-key-123"
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have error: %v", err)
	}

	if cf.Comm.SSHTemporaryKeyPairName != "ssh-key-123" {
		t.Fatalf("invalid ssh key pair value: %v", cf.Comm.SSHTemporaryKeyPairName)
	}
}

func TestTencentCloudRunConfigPrepare_SSHPrivateIp(t *testing.T) {
	cf := testConfig()
	if cf.SSHPrivateIp != false {
		t.Fatalf("invalid ssh_private_ip value: %v", cf.SSHPrivateIp)
	}
	cf.SSHPrivateIp = true
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have error: %v", err)
	}
	if cf.SSHPrivateIp != true {
		t.Fatalf("invalud ssh_private_ip value: %v", cf.SSHPrivateIp)
	}
}

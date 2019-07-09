package jdcloud

import (
	"testing"
)

func TestJDCloudInstanceSpecConfig_Prepare(t *testing.T) {

	specs := &JDCloudInstanceSpecConfig{}
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when there's nothing set")
	}

	specs.InstanceName = "packer_test_instance_name"
	specs.InstanceType = "packer_test_instance_type"
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when base-image not given")
	}

	specs.ImageId = "img-packer-test"
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when credentials not set")
	}

	specs.Comm.SSHPassword = "abc123"
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when username = nil")
	}

	specs.Comm.SSHUsername = "root"
	if err := specs.Prepare(nil); err != nil {
		t.Fatalf("Test shouldn't fail when password set ")
	}

	specs.Comm.SSHPassword = ""
	specs.Comm.SSHTemporaryKeyPairName = "abc"
	if err := specs.Prepare(nil); err != nil {
		t.Fatalf("Test shouldn't fail when temp password set")
	}

	specs.Comm.SSHTemporaryKeyPairName = ""
	specs.Comm.SSHPrivateKeyFile = "abc"
	specs.Comm.SSHKeyPairName = ""
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when SSHKeypairName missing")
	}

	specs.Comm.SSHPrivateKeyFile = "abc"
	specs.Comm.SSHKeyPairName = "123"
	if err := specs.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when private key pair path is wrong ")
	}

	specs.Comm.SSHPrivateKeyFile = "/Users/mac/.ssh/id_rsa"
	if err := specs.Prepare(nil); err != nil {
		t.Fatalf("Test shouldn't fail when private everything is given properly")
	}
}

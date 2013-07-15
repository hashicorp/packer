package ebs

import (
	"github.com/mitchellh/packer/packer"
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

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":    "foo",
		"secret_key":    "bar",
		"source_ami":    "foo",
		"instance_type": "foo",
		"region":        "us-east-1",
		"ssh_username":  "root",
		"ami_name":      "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	err := b.Prepare(c)
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_AMIName(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["ami_name"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ami_name"] = "foo {{"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "ami_name")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_InstanceType(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["instance_type"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.InstanceType != "foo" {
		t.Errorf("invalid: %s", b.config.InstanceType)
	}

	// Test bad
	delete(config, "instance_type")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Region(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["region"] = "us-east-1"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Region != "us-east-1" {
		t.Errorf("invalid: %s", b.config.Region)
	}

	// Test bad
	delete(config, "region")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test invalid
	config["region"] = "i-am-not-real"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_SourceAmi(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["source_ami"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SourceAmi != "foo" {
		t.Errorf("invalid: %s", b.config.SourceAmi)
	}

	// Test bad
	delete(config, "source_ami")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_SSHPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHPort != 22 {
		t.Errorf("invalid: %d", b.config.SSHPort)
	}

	// Test set
	config["ssh_port"] = 35
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHPort != 35 {
		t.Errorf("invalid: %d", b.config.SSHPort)
	}
}

func TestBuilderPrepare_SSHTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with a bad value
	config["ssh_timeout"] = "this is not good"
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["ssh_timeout"] = "5s"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHUsername(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["ssh_username"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHUsername != "foo" {
		t.Errorf("invalid: %s", b.config.SSHUsername)
	}

	// Test bad
	delete(config, "ssh_username")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

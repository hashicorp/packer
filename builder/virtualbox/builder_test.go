package virtualbox

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_md5":      "foo",
		"iso_url":      "http://www.google.com/",
		"ssh_username": "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Error("Builder must implement builder.")
	}
}

func TestBuilderPrepare_Defaults(t *testing.T) {
	var b Builder
	config := testConfig()
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestOSType != "Other" {
		t.Errorf("bad guest OS type: %s", b.config.GuestOSType)
	}

	if b.config.OutputDir != "virtualbox" {
		t.Errorf("bad output dir: %s", b.config.OutputDir)
	}

	if b.config.SSHHostPortMin != 2222 {
		t.Errorf("bad min ssh host port: %d", b.config.SSHHostPortMin)
	}

	if b.config.SSHHostPortMax != 4444 {
		t.Errorf("bad max ssh host port: %d", b.config.SSHHostPortMax)
	}

	if b.config.SSHPort != 22 {
		t.Errorf("bad ssh port: %d", b.config.SSHPort)
	}

	if b.config.VMName != "packer" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}
}

func TestBuilderPrepare_BootWait(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with a bad boot_wait
	config["boot_wait"] = "this is not good"
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["boot_wait"] = "5s"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_HTTPPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["http_port_min"] = 1000
	config["http_port_max"] = 500
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["http_port_min"] = -500
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["http_port_min"] = 500
	config["http_port_max"] = 1000
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ISOMD5(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["iso_md5"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_md5"] = "FOo"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOMD5 != "foo" {
		t.Fatalf("should've lowercased: %s", b.config.ISOMD5)
	}
}

func TestBuilderPrepare_ISOUrl(t *testing.T) {
	var b Builder
	config := testConfig()

	config["iso_url"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["iso_url"] = "i/am/a/file/that/doesnt/exist"
	err = b.Prepare(config)
	if err == nil {
		t.Error("should have error")
	}

	config["iso_url"] = "file:i/am/a/file/that/doesnt/exist"
	err = b.Prepare(config)
	if err == nil {
		t.Error("should have error")
	}

	config["iso_url"] = "http://www.packer.io"
	err = b.Prepare(config)
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["iso_url"] = tf.Name()
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOUrl != "file://"+tf.Name() {
		t.Fatalf("iso_url should be modified: %s", b.config.ISOUrl)
	}
}

func TestBuilderPrepare_SSHHostPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["ssh_host_port_min"] = 1000
	config["ssh_host_port_max"] = 500
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["ssh_host_port_min"] = -500
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["ssh_host_port_min"] = 500
	config["ssh_host_port_max"] = 1000
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHUser(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ssh_username"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["ssh_username"] = "exists"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHWaitTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with a bad value
	config["ssh_wait_timeout"] = "this is not good"
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["ssh_wait_timeout"] = "5s"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

package vmware

import (
	"github.com/mitchellh/packer/packer"
	"testing"
	"time"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_url":      "foo",
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

func TestBuilderPrepare_Defaults(t *testing.T) {
	var b Builder
	config := testConfig()
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DiskName != "disk" {
		t.Errorf("bad disk name: %s", b.config.DiskName)
	}

	if b.config.OutputDir != "vmware" {
		t.Errorf("bad output dir: %s", b.config.OutputDir)
	}

	if b.config.SSHWaitTimeout != (20 * time.Minute) {
		t.Errorf("bad wait timeout: %s", b.config.SSHWaitTimeout)
	}

	if b.config.VMName != "packer" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}
}

func TestBuilderPrepare_ISOUrl(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test iso_url
	config["iso_url"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["iso_url"] = "exists"
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ShutdownTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with a bad value
	config["shutdown_timeout"] = "this is not good"
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["shutdown_timeout"] = "5s"
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

func TestBuilderPrepare_VNCPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["vnc_port_min"] = 1000
	config["vnc_port_max"] = 500
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["vnc_port_min"] = -500
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["vnc_port_min"] = 500
	config["vnc_port_max"] = 1000
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

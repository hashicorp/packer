package iso

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_checksum":           "foo",
		"iso_checksum_type":      "md5",
		"iso_url":                "http://www.google.com/",
		"shutdown_command":       "yes",
		"ssh_username":           "foo",
		"parallels_tools_flavor": "lin",

		packer.BuildNameConfigKey: "foo",
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
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestOSType != "other" {
		t.Errorf("bad guest OS type: %s", b.config.GuestOSType)
	}

	if b.config.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}
}

func TestBuilderPrepare_DiskSize(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "disk_size")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.DiskSize != 40000 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}

	config["disk_size"] = 60000
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DiskSize != 60000 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}
}

func TestBuilderPrepare_GuestOSType(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "guest_os_distribution")

	// Test deprecated parameter
	config["guest_os_distribution"] = "bolgenos"
	warns, err := b.Prepare(config)
	if len(warns) == 0 {
		t.Fatalf("should have warning")
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if b.config.GuestOSType != "bolgenos" {
		t.Fatalf("bad: %s", b.config.GuestOSType)
	}
}

func TestBuilderPrepare_HardDriveInterface(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "hard_drive_interface")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.HardDriveInterface != "sata" {
		t.Fatalf("bad: %s", b.config.HardDriveInterface)
	}

	// Test with a bad
	config["hard_drive_interface"] = "fake"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good
	config["hard_drive_interface"] = "scsi"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ParallelsToolsHostPath(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "parallels_tools_host_path")

	// Test that it is deprecated
	config["parallels_tools_host_path"] = "/path/to/iso"
	warns, err := b.Prepare(config)
	if len(warns) == 0 {
		t.Fatalf("should have warning")
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

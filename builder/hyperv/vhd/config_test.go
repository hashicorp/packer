package vhd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"source_path":             "disk.vhd",
		"shutdown_command":        "yes",
		"ssh_username":            "foo",
		"ram_size":                64,
		"disk_size":               256,
		"guest_additions_mode":    "none",
		packer.BuildNameConfigKey: "foo",
	}
}

func TestConfig_Defaults(t *testing.T) {
	conf, warns, err := NewConfig(testConfig())
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if conf.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", conf.VMName)
	}
}

func TestConfig_DiskSize(t *testing.T) {
	config := testConfig()

	delete(config, "disk_size")
	conf, warns, err := NewConfig(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if conf.DiskSize != 40*1024 {
		t.Fatalf("bad size: %d", conf.DiskSize)
	}

	config["disk_size"] = 256
	conf, warns, err = NewConfig(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if conf.DiskSize != 256 {
		t.Fatalf("bad size: %d", conf.DiskSize)
	}
}

func TestConfig_FloppyFiles(t *testing.T) {
	config := testConfig()

	delete(config, "floppy_files")
	conf, warns, err := NewConfig(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if len(conf.FloppyFiles) != 0 {
		t.Fatalf("bad: %#v", conf.FloppyFiles)
	}

	floppiesPath := "../../../common/test-fixtures/floppies"
	config["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppiesPath), fmt.Sprintf("%s/foo.ps1", floppiesPath)}
	conf, warns, err = NewConfig(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{fmt.Sprintf("%s/bar.bat", floppiesPath), fmt.Sprintf("%s/foo.ps1", floppiesPath)}
	if !reflect.DeepEqual(conf.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", conf.FloppyFiles)
	}
}

func TestConfig_InvalidFloppies(t *testing.T) {
	config := testConfig()
	config["floppy_files"] = []string{"nonexistent.bat", "nonexistent.ps1"}
	_, _, errs := NewConfig(config)
	if errs == nil {
		t.Fatalf("Nonexistent floppies should trigger multierror")
	}

	if len(errs.(*packer.MultiError).Errors) != 2 {
		t.Fatalf("Multierror should work and report 2 errors")
	}
}

func TestConfig_InvalidKey(t *testing.T) {
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warns, err := NewConfig(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

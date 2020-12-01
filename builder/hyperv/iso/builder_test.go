// +build !windows

package iso

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	hypervcommon "github.com/hashicorp/packer/builder/hyperv/common"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_checksum":            "md5:0B0F137F17AC10944716020B018F8126",
		"iso_url":                 "http://www.packer.io",
		"shutdown_command":        "yes",
		"ssh_username":            "foo",
		"memory":                  64,
		"disk_size":               256,
		"disk_block_size":         1,
		"guest_additions_mode":    "none",
		"disk_additional_size":    "50000,40000,30000",
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

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}
}

func TestBuilderPrepare_DiskSize(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "disk_size")
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.DiskSize != 40*1024 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}

	config["disk_size"] = 256
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DiskSize != 256 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}
}

func TestBuilderPrepare_DiskBlockSize(t *testing.T) {
	var b Builder
	config := testConfig()
	expected_default_block_size := uint(32)
	expected_min_block_size := uint(0)
	expected_max_block_size := uint(256)

	// Test default with empty disk_block_size
	delete(config, "disk_block_size")
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}
	if b.config.DiskBlockSize != expected_default_block_size {
		t.Fatalf("bad default block size with empty config: %d. Expected %d", b.config.DiskBlockSize,
			expected_default_block_size)
	}

	test_sizes := []uint{0, 1, 32, 256, 512, 1 * 1024, 32 * 1024}
	for _, test_size := range test_sizes {
		config["disk_block_size"] = test_size
		b = Builder{}
		_, warns, err = b.Prepare(config)
		if test_size > expected_max_block_size || test_size < expected_min_block_size {
			if len(warns) > 0 {
				t.Fatalf("bad, should have no warns: %#v", warns)
			}
			if err == nil {
				t.Fatalf("bad, should have error. disk_block_size=%d outside expected valid range [%d,%d]",
					test_size, expected_min_block_size, expected_max_block_size)
			}
		} else {
			if len(warns) > 0 {
				t.Fatalf("bad: %#v", warns)
			}
			if err != nil {
				t.Fatalf("bad, should not have error: %s", err)
			}
			if test_size == 0 {
				if b.config.DiskBlockSize != expected_default_block_size {
					t.Fatalf("bad default block size with 0 value config: %d. Expected: %d",
						b.config.DiskBlockSize, expected_default_block_size)
				}
			} else {
				if b.config.DiskBlockSize != test_size {
					t.Fatalf("bad block size with 0 value config: %d. Expected: %d", b.config.DiskBlockSize,
						expected_default_block_size)
				}
			}
		}
	}
}

func TestBuilderPrepare_FixedVHDFormat(t *testing.T) {
	var b Builder
	config := testConfig()
	config["use_fixed_vhd_format"] = true
	config["generation"] = 1
	config["skip_compaction"] = true
	config["differencing_disk"] = false

	// use_fixed_vhd_format should work with generation = 1, skip_compaction
	// = true, and differencing_disk = false
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	//use_fixed_vhd_format should not work with differencing_disk = true
	config["differencing_disk"] = true
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
	config["differencing_disk"] = false

	//use_fixed_vhd_format should not work with skip_compaction = false
	config["skip_compaction"] = false
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
	config["skip_compaction"] = true

	//use_fixed_vhd_format should not work with generation = 2
	config["generation"] = 2
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_FloppyFiles(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "floppy_files")
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if len(b.config.FloppyFiles) != 0 {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}

	floppiesPath := "../../../packer-plugin-sdk/test-fixtures/floppies"
	config["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppiesPath), fmt.Sprintf("%s/foo.ps1", floppiesPath)}
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{fmt.Sprintf("%s/bar.bat", floppiesPath), fmt.Sprintf("%s/foo.ps1", floppiesPath)}
	if !reflect.DeepEqual(b.config.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}
}

func TestBuilderPrepare_InvalidFloppies(t *testing.T) {
	var b Builder
	config := testConfig()
	config["floppy_files"] = []string{"nonexistent.bat", "nonexistent.ps1"}
	b = Builder{}
	_, _, errs := b.Prepare(config)
	if errs == nil {
		t.Fatalf("Nonexistent floppies should trigger multierror")
	}

	if len(errs.(*packersdk.MultiError).Errors) != 2 {
		t.Fatalf("Multierror should work and report 2 errors")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ISOChecksum(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["iso_checksum"] = ""
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test invalid checksum
	config["iso_checksum"] = "FOo"
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum"] = "md5:0B0F137F17AC10944716020B018F8126"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

}

func TestBuilderPrepare_ISOChecksumType(t *testing.T) {
	var b Builder
	config := testConfig()

	config["iso_checksum"] = "0B0F137F17AC10944716020B018F8126"
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test good
	config["iso_checksum"] = "mD5:0B0F137F17AC10944716020B018F8126"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test unknown
	config["iso_checksum"] = "fake:foo"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Log("should error in prepare but go-getter doesn't let us validate yet. This will fail before dl.")
	}

	// Test none
	config["iso_checksum"] = "none"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) == 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ISOUrl(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "iso_url")
	delete(config, "iso_urls")

	// Test both empty
	config["iso_url"] = ""
	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test iso_url set
	config["iso_url"] = "http://www.packer.io"
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected := []string{"http://www.packer.io"}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}

	// Test both set
	config["iso_url"] = "http://www.packer.io"
	config["iso_urls"] = []string{"http://www.packer.io"}
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test just iso_urls set
	delete(config, "iso_url")
	config["iso_urls"] = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}

	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}
}

func TestBuilderPrepare_SizeNotRequiredWhenUsingExistingHarddrive(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "iso_url")
	delete(config, "iso_urls")
	delete(config, "disk_size")

	config["disk_size"] = 1

	// Test just iso_urls set but with vhdx
	delete(config, "iso_url")
	config["iso_urls"] = []string{
		"http://www.packer.io/hdd.vhdx",
		"http://www.hashicorp.com/dvd.iso",
	}

	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected := []string{
		"http://www.packer.io/hdd.vhdx",
		"http://www.hashicorp.com/dvd.iso",
	}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}

	// Test just iso_urls set but with vhd
	delete(config, "iso_url")
	config["iso_urls"] = []string{
		"http://www.packer.io/hdd.vhd",
		"http://www.hashicorp.com/dvd.iso",
	}

	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected = []string{
		"http://www.packer.io/hdd.vhd",
		"http://www.hashicorp.com/dvd.iso",
	}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}
}

func TestBuilderPrepare_SizeIsRequiredWhenNotUsingExistingHarddrive(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "iso_url")
	delete(config, "iso_urls")
	delete(config, "disk_size")

	config["disk_size"] = 1

	// Test just iso_urls set but with vhdx
	delete(config, "iso_url")
	config["iso_urls"] = []string{
		"http://www.packer.io/os.iso",
		"http://www.hashicorp.com/dvd.iso",
	}

	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Errorf("should have error")
	}

	expected := []string{
		"http://www.packer.io/os.iso",
		"http://www.hashicorp.com/dvd.iso",
	}
	if !reflect.DeepEqual(b.config.ISOUrls, expected) {
		t.Fatalf("bad: %#v", b.config.ISOUrls)
	}
}

func TestBuilderPrepare_MaximumOfSixtyFourAdditionalDisks(t *testing.T) {
	var b Builder
	config := testConfig()

	disks := make([]string, 65)
	for i := range disks {
		disks[i] = strconv.Itoa(i)
	}
	config["disk_additional_size"] = disks

	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Errorf("should have error")
	}

}

func TestBuilderPrepare_CommConfig(t *testing.T) {
	// Test Winrm
	{
		config := testConfig()
		config["communicator"] = "winrm"
		config["winrm_username"] = "username"
		config["winrm_password"] = "password"
		config["winrm_host"] = "1.2.3.4"

		var b Builder
		_, warns, err := b.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("bad: %#v", warns)
		}
		if err != nil {
			t.Fatalf("should not have error: %s", err)
		}

		if b.config.Comm.WinRMUser != "username" {
			t.Errorf("bad winrm_username: %s", b.config.Comm.WinRMUser)
		}
		if b.config.Comm.WinRMPassword != "password" {
			t.Errorf("bad winrm_password: %s", b.config.Comm.WinRMPassword)
		}
		if host := b.config.Comm.Host(); host != "1.2.3.4" {
			t.Errorf("bad host: %s", host)
		}
	}

	// Test SSH
	{
		config := testConfig()
		config["communicator"] = "ssh"
		config["ssh_username"] = "username"
		config["ssh_password"] = "password"
		config["ssh_host"] = "1.2.3.4"

		var b Builder
		_, warns, err := b.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("bad: %#v", warns)
		}
		if err != nil {
			t.Fatalf("should not have error: %s", err)
		}

		if b.config.Comm.SSHUsername != "username" {
			t.Errorf("bad ssh_username: %s", b.config.Comm.SSHUsername)
		}
		if b.config.Comm.SSHPassword != "password" {
			t.Errorf("bad ssh_password: %s", b.config.Comm.SSHPassword)
		}
		if host := b.config.Comm.Host(); host != "1.2.3.4" {
			t.Errorf("bad host: %s", host)
		}
	}

}

func TestUserVariablesInBootCommand(t *testing.T) {
	var b Builder
	config := testConfig()

	config[packer.UserVariablesConfigKey] = map[string]string{"test-variable": "test"}
	config["boot_command"] = []string{"blah {{user `test-variable`}} blah"}

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	ui := packer.TestUi(t)
	hook := &packersdk.MockHook{}
	driver := &hypervcommon.DriverMock{}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("http_port", 0)
	state.Put("http_ip", "0.0.0.0")
	state.Put("ui", ui)
	state.Put("vmName", "packer-foo")

	step := &hypervcommon.StepTypeBootCommand{
		BootCommand: b.config.FlatBootCommand(),
		SwitchName:  b.config.SwitchName,
		Ctx:         b.config.ctx,
	}

	ret := step.Run(context.Background(), state)
	if ret != multistep.ActionContinue {
		t.Fatalf("should not have error: %#v", ret)
	}
}

func TestBuilderPrepare_UseLegacyNetworkAdapter(t *testing.T) {
	var b Builder
	config := testConfig()

	// should be allowed for default config
	config["use_legacy_network_adapter"] = true

	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	// should not be allowed for gen 2
	config["generation"] = 2

	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

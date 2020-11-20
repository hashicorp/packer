package vmcx

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"io/ioutil"
	"os"

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
		"switch_name":             "switch", // to avoid using builder.detectSwitchName which can lock down in travis-ci
		"memory":                  64,
		"guest_additions_mode":    "none",
		"clone_from_vmcx_path":    "generated",
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

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

func TestBuilderPrepare_CloneFromExistingMachineOrImportFromExportedMachineSettingsRequired(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "clone_from_vmcx_path")

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ExportedMachinePathDoesNotExist(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	//Delete the folder immediately
	os.RemoveAll(td)

	config["clone_from_vmcx_path"] = td

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ExportedMachinePathExists(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	//Only delete afterwards
	defer os.RemoveAll(td)

	config["clone_from_vmcx_path"] = td

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func disabled_TestBuilderPrepare_CloneFromVmSettingUsedSoNoCloneFromVmcxPathRequired(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "clone_from_vmcx_path")

	config["clone_from_vm_name"] = "test_machine_name_that_does_not_exist"

	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}

	if err == nil {
		t.Fatal("should have error")
	} else {
		errorMessage := err.Error()
		if errorMessage != "1 error(s) occurred:\n\n* Virtual machine 'test_machine_name_that_does_not_exist' "+
			"to clone from does not exist." {
			t.Fatalf("should not have error: %s", err)
		}
	}
}

func TestBuilderPrepare_ISOChecksum(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

	// Test bad
	config["iso_checksum"] = ""
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum"] = "0B0F137F17AC10944716020B018F8126"
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

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

	delete(config, "iso_url")
	delete(config, "iso_urls")

	// Test both empty (should be allowed, as we cloning a vm so we probably don't need an ISO file)
	config["iso_url"] = ""
	b = Builder{}
	_, warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatal("should not have an error")
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

func TestBuilderPrepare_FloppyFiles(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

	floppies_path := "../../../packer-plugin-sdk/test-fixtures/floppies"
	config["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	b = Builder{}
	_, warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	if !reflect.DeepEqual(b.config.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}
}

func TestBuilderPrepare_InvalidFloppies(t *testing.T) {
	var b Builder
	config := testConfig()

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

func TestBuilderPrepare_CommConfig(t *testing.T) {
	// Test Winrm
	{
		config := testConfig()

		//Create vmcx folder
		td, err := ioutil.TempDir("", "packer")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		defer os.RemoveAll(td)
		config["clone_from_vmcx_path"] = td

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

		//Create vmcx folder
		td, err := ioutil.TempDir("", "packer")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		defer os.RemoveAll(td)
		config["clone_from_vmcx_path"] = td

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

	//Create vmcx folder
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)
	config["clone_from_vmcx_path"] = td

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

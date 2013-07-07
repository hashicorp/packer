package virtualbox

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_md5":      "foo",
		"iso_url":      "http://www.google.com/",
		"ssh_username": "foo",

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
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestOSType != "Other" {
		t.Errorf("bad guest OS type: %s", b.config.GuestOSType)
	}

	if b.config.OutputDir != "output-foo" {
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

	if b.config.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}
}

func TestBuilderPrepare_BootWait(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "boot_wait")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.RawBootWait != "10s" {
		t.Fatalf("bad value: %s", b.config.RawBootWait)
	}

	// Test with a bad boot_wait
	config["boot_wait"] = "this is not good"
	err = b.Prepare(config)
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

func TestBuilderPrepare_DiskSize(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "disk_size")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.DiskSize != 40000 {
		t.Fatalf("bad size: %d", b.config.DiskSize)
	}

	config["disk_size"] = 60000
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DiskSize != 60000 {
		t.Fatalf("bad size: %s", b.config.DiskSize)
	}
}

func TestBuilderPrepare_GuestAdditionsPath(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "guest_additions_path")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.GuestAdditionsPath != "VBoxGuestAdditions.iso" {
		t.Fatalf("bad: %s", b.config.GuestAdditionsPath)
	}

	config["guest_additions_path"] = "foo"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestAdditionsPath != "foo" {
		t.Fatalf("bad size: %s", b.config.GuestAdditionsPath)
	}
}

func TestBuilderPrepare_GuestAdditionsSHA256(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "guest_additions_sha256")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if b.config.GuestAdditionsSHA256 != "" {
		t.Fatalf("bad: %s", b.config.GuestAdditionsSHA256)
	}

	config["guest_additions_sha256"] = "FOO"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestAdditionsSHA256 != "foo" {
		t.Fatalf("bad size: %s", b.config.GuestAdditionsSHA256)
	}
}

func TestBuilderPrepare_GuestAdditionsURL(t *testing.T) {
	var b Builder
	config := testConfig()

	config["guest_additions_url"] = ""
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.GuestAdditionsURL != "" {
		t.Fatalf("should be empty: %s", b.config.GuestAdditionsURL)
	}

	config["guest_additions_url"] = "i/am/a/file/that/doesnt/exist"
	err = b.Prepare(config)
	if err == nil {
		t.Error("should have error")
	}

	config["guest_additions_url"] = "file:i/am/a/file/that/doesnt/exist"
	err = b.Prepare(config)
	if err == nil {
		t.Error("should have error")
	}

	config["guest_additions_url"] = "http://www.packer.io"
	err = b.Prepare(config)
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["guest_additions_url"] = tf.Name()
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestAdditionsURL != "file://"+tf.Name() {
		t.Fatalf("guest_additions_url should be modified: %s", b.config.GuestAdditionsURL)
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

func TestBuilderPrepare_OutputDir(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with existing dir
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	config["output_directory"] = dir
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["output_directory"] = "i-hope-i-dont-exist"
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

	// Test a default boot_wait
	delete(config, "ssh_wait_timeout")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.RawSSHWaitTimeout != "20m" {
		t.Fatalf("bad value: %s", b.config.RawSSHWaitTimeout)
	}

	// Test with a bad value
	config["ssh_wait_timeout"] = "this is not good"
	err = b.Prepare(config)
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

func TestBuilderPrepare_VBoxManage(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with empty
	delete(config, "vboxmanage")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(b.config.VBoxManage, [][]string{}) {
		t.Fatalf("bad: %#v", b.config.VBoxManage)
	}

	// Test with a good one
	config["vboxmanage"] = [][]interface{}{
		[]interface{}{"foo", "bar", "baz"},
	}

	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := [][]string{
		[]string{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(b.config.VBoxManage, expected) {
		t.Fatalf("bad: %#v", b.config.VBoxManage)
	}
}

func TestBuilderPrepare_VBoxVersionFile(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test empty
	delete(config, "virtualbox_version_file")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.VBoxVersionFile != ".vbox_version" {
		t.Fatalf("bad value: %s", b.config.VBoxVersionFile)
	}

	// Test with a good one
	config["virtualbox_version_file"] = "foo"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.VBoxVersionFile != "foo" {
		t.Fatalf("bad value: %s", b.config.VBoxVersionFile)
	}
}

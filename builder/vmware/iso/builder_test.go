package iso

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_checksum":      "foo",
		"iso_checksum_type": "md5",
		"iso_url":           "http://www.packer.io",
		"shutdown_command":  "foo",
		"ssh_username":      "foo",

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

func TestBuilderPrepare_ISOChecksum(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["iso_checksum"] = ""
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum"] = "FOo"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOChecksum != "foo" {
		t.Fatalf("should've lowercased: %s", b.config.ISOChecksum)
	}
}

func TestBuilderPrepare_ISOChecksumType(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["iso_checksum_type"] = ""
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum_type"] = "mD5"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOChecksumType != "md5" {
		t.Fatalf("should've lowercased: %s", b.config.ISOChecksumType)
	}

	// Test unknown
	config["iso_checksum_type"] = "fake"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test none
	config["iso_checksum_type"] = "none"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) == 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOChecksumType != "none" {
		t.Fatalf("should've lowercased: %s", b.config.ISOChecksumType)
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

	if b.config.DiskName != "disk" {
		t.Errorf("bad disk name: %s", b.config.DiskName)
	}

	if b.config.OutputDir != "output-foo" {
		t.Errorf("bad output dir: %s", b.config.OutputDir)
	}

	if b.config.SSHWaitTimeout != (20 * time.Minute) {
		t.Errorf("bad wait timeout: %s", b.config.SSHWaitTimeout)
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
		t.Fatalf("bad size: %s", b.config.DiskSize)
	}
}

func TestBuilderPrepare_FloppyFiles(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "floppy_files")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if len(b.config.FloppyFiles) != 0 {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}

	config["floppy_files"] = []string{"foo", "bar"}
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual(b.config.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}
}

func TestBuilderPrepare_HTTPPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["http_port_min"] = 1000
	config["http_port_max"] = 500
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["http_port_min"] = -500
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["http_port_min"] = 500
	config["http_port_max"] = 1000
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

func TestBuilderPrepare_ISOUrl(t *testing.T) {
	var b Builder
	config := testConfig()
	delete(config, "iso_url")
	delete(config, "iso_urls")

	// Test both epty
	config["iso_url"] = ""
	b = Builder{}
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test iso_url set
	config["iso_url"] = "http://www.packer.io"
	b = Builder{}
	warns, err = b.Prepare(config)
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
	warns, err = b.Prepare(config)
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
	warns, err = b.Prepare(config)
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
	b = Builder{}
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["output_directory"] = "i-hope-i-dont-exist"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ToolsUploadPath(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default
	delete(config, "tools_upload_path")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.ToolsUploadPath == "" {
		t.Fatalf("bad value: %s", b.config.ToolsUploadPath)
	}

	// Test with a bad value
	config["tools_upload_path"] = "{{{nope}"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["tools_upload_path"] = "hey"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_VMXTemplatePath(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["vmx_template_path"] = "/i/dont/exist/forreal"
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err := tf.Write([]byte("HELLO!")); err != nil {
		t.Fatalf("err: %s", err)
	}

	config["vmx_template_path"] = tf.Name()
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Bad template
	tf2, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf2.Name())
	defer tf2.Close()

	if _, err := tf2.Write([]byte("{{foo}")); err != nil {
		t.Fatalf("err: %s", err)
	}

	config["vmx_template_path"] = tf2.Name()
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_VNCPort(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["vnc_port_min"] = 1000
	config["vnc_port_max"] = 500
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["vnc_port_min"] = -500
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["vnc_port_min"] = 500
	config["vnc_port_max"] = 1000
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

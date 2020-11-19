package qemu

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

var testPem = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAxd4iamvrwRJvtNDGQSIbNvvIQN8imXTRWlRY62EvKov60vqu
hh+rDzFYAIIzlmrJopvOe0clqmi3mIP9dtkjPFrYflq52a2CF5q+BdwsJXuRHbJW
LmStZUwW1khSz93DhvhmK50nIaczW63u4EO/jJb3xj+wxR1Nkk9bxi3DDsYFt8SN
AzYx9kjlEYQ/+sI4/ATfmdV9h78SVotjScupd9KFzzi76gWq9gwyCBLRynTUWlyD
2UOfJRkOvhN6/jKzvYfVVwjPSfA9IMuooHdScmC4F6KBKJl/zf/zETM0XyzIDNmH
uOPbCiljq2WoRM+rY6ET84EO0kVXbfx8uxUsqQIDAQABAoIBAQCkPj9TF0IagbM3
5BSs/CKbAWS4dH/D4bPlxx4IRCNirc8GUg+MRb04Xz0tLuajdQDqeWpr6iLZ0RKV
BvreLF+TOdV7DNQ4XE4gSdJyCtCaTHeort/aordL3l0WgfI7mVk0L/yfN1PEG4YG
E9q1TYcyrB3/8d5JwIkjabxERLglCcP+geOEJp+QijbvFIaZR/n2irlKW4gSy6ko
9B0fgUnhkHysSg49ChHQBPQ+o5BbpuLrPDFMiTPTPhdfsvGGcyCGeqfBA56oHcSF
K02Fg8OM+Bd1lb48LAN9nWWY4WbwV+9bkN3Ym8hO4c3a/Dxf2N7LtAQqWZzFjvM3
/AaDvAgBAoGBAPLD+Xn1IYQPMB2XXCXfOuJewRY7RzoVWvMffJPDfm16O7wOiW5+
2FmvxUDayk4PZy6wQMzGeGKnhcMMZTyaq2g/QtGfrvy7q1Lw2fB1VFlVblvqhoJa
nMJojjC4zgjBkXMHsRLeTmgUKyGs+fdFbfI6uejBnnf+eMVUMIdJ+6I9AoGBANCn
kWO9640dttyXURxNJ3lBr2H3dJOkmD6XS+u+LWqCSKQe691Y/fZ/ZL0Oc4Mhy7I6
hsy3kDQ5k2V0fkaNODQIFJvUqXw2pMewUk8hHc9403f4fe9cPrL12rQ8WlQw4yoC
v2B61vNczCCUDtGxlAaw8jzSRaSI5s6ax3K7enbdAoGBAJB1WYDfA2CoAQO6y9Sl
b07A/7kQ8SN5DbPaqrDrBdJziBQxukoMJQXJeGFNUFD/DXFU5Fp2R7C86vXT7HIR
v6m66zH+CYzOx/YE6EsUJms6UP9VIVF0Rg/RU7teXQwM01ZV32LQ8mswhTH20o/3
uqMHmxUMEhZpUMhrfq0isyApAoGAe1UxGTXfj9AqkIVYylPIq2HqGww7+jFmVEj1
9Wi6S6Sq72ffnzzFEPkIQL/UA4TsdHMnzsYKFPSbbXLIWUeMGyVTmTDA5c0e5XIR
lPhMOKCAzv8w4VUzMnEkTzkFY5JqFCD/ojW57KvDdNZPVB+VEcdxyAW6aKELXMAc
eHLc1nkCgYEApm/motCTPN32nINZ+Vvywbv64ZD+gtpeMNP3CLrbe1X9O+H52AXa
1jCoOldWR8i2bs2NVPcKZgdo6fFULqE4dBX7Te/uYEIuuZhYLNzRO1IKU/YaqsXG
3bfQ8hKYcSnTfE0gPtLDnqCIxTocaGLSHeG3TH9fTw+dA8FvWpUztI4=
-----END RSA PRIVATE KEY-----
`

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"iso_checksum":            "md5:0B0F137F17AC10944716020B018F8126",
		"iso_url":                 "http://www.google.com/",
		"ssh_username":            "foo",
		packer.BuildNameConfigKey: "foo",
	}
}

func TestBuilderPrepare_Defaults(t *testing.T) {
	var c Config
	config := testConfig()
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if c.OutputDir != "output-foo" {
		t.Errorf("bad output dir: %s", c.OutputDir)
	}

	if c.CommConfig.HostPortMin != 2222 {
		t.Errorf("bad min ssh host port: %d", c.CommConfig.HostPortMin)
	}

	if c.CommConfig.HostPortMax != 4444 {
		t.Errorf("bad max ssh host port: %d", c.CommConfig.HostPortMax)
	}

	if c.CommConfig.Comm.SSHPort != 22 {
		t.Errorf("bad ssh port: %d", c.CommConfig.Comm.SSHPort)
	}

	if c.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", c.VMName)
	}

	if c.Format != "qcow2" {
		t.Errorf("bad format: %s", c.Format)
	}
}

func TestBuilderPrepare_VNCBindAddress(t *testing.T) {
	var c Config
	config := testConfig()

	// Test a default boot_wait
	delete(config, "vnc_bind_address")
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if c.VNCBindAddress != "127.0.0.1" {
		t.Fatalf("bad value: %s", c.VNCBindAddress)
	}
}

func TestBuilderPrepare_DiskCompaction(t *testing.T) {
	var c Config
	config := testConfig()

	// Bad
	config["skip_compaction"] = false
	config["disk_compression"] = true
	config["format"] = "img"
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
	if c.SkipCompaction != true {
		t.Fatalf("SkipCompaction should be true")
	}
	if c.DiskCompression != false {
		t.Fatalf("DiskCompression should be false")
	}

	// Good
	config["skip_compaction"] = false
	config["disk_compression"] = true
	config["format"] = "qcow2"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if c.SkipCompaction != false {
		t.Fatalf("SkipCompaction should be false")
	}
	if c.DiskCompression != true {
		t.Fatalf("DiskCompression should be true")
	}
}

func TestBuilderPrepare_DiskSize(t *testing.T) {
	type testcase struct {
		InputSize   string
		OutputSize  string
		ErrExpected bool
	}

	testCases := []testcase{
		{"", "40960M", false},       // not provided
		{"12345", "12345M", false},  // no unit given, defaults to M
		{"12345x", "12345x", true},  // invalid unit
		{"12345T", "12345T", false}, // terabytes
		{"12345b", "12345b", false}, // bytes get preserved when set.
		{"60000M", "60000M", false}, // Original test case
	}
	for _, tc := range testCases {
		// Set input disk size
		var c Config
		config := testConfig()
		delete(config, "disk_size")
		config["disk_size"] = tc.InputSize

		warns, err := c.Prepare(config)
		if len(warns) > 0 {
			t.Fatalf("bad: %#v", warns)
		}
		if (err == nil) == tc.ErrExpected {
			t.Fatalf("bad: error when providing disk size %s; Err expected: %t; err recieved: %v", tc.InputSize, tc.ErrExpected, err)
		}

		if c.DiskSize != tc.OutputSize {
			t.Fatalf("bad size: received: %s but expected %s", c.DiskSize, tc.OutputSize)
		}
	}
}

func TestBuilderPrepare_AdditionalDiskSize(t *testing.T) {
	var c Config
	config := testConfig()

	config["disk_additional_size"] = []string{"1M"}
	config["disk_image"] = true
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should have error")
	}

	delete(config, "disk_image")
	config["disk_additional_size"] = []string{"1M"}
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if c.AdditionalDiskSize[0] != "1M" {
		t.Fatalf("bad size: %s", c.AdditionalDiskSize)
	}
}

func TestBuilderPrepare_Format(t *testing.T) {
	var c Config
	config := testConfig()

	// Bad
	config["format"] = "illegal value"
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["format"] = "qcow2"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Good
	config["format"] = "raw"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_UseBackingFile(t *testing.T) {
	var c Config
	config := testConfig()

	config["use_backing_file"] = true

	// Bad: iso_url is not a disk_image
	config["disk_image"] = false
	config["format"] = "qcow2"
	c = Config{}
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad: format is not 'qcow2'
	config["disk_image"] = true
	config["format"] = "raw"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good: iso_url is a disk image and format is 'qcow2'
	config["disk_image"] = true
	config["format"] = "qcow2"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SkipResizeDisk(t *testing.T) {
	config := testConfig()
	config["skip_resize_disk"] = true
	config["disk_image"] = false

	var c Config
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Errorf("unexpected warns when calling prepare with skip_resize_disk set to true: %#v", warns)
	}
	if err == nil {
		t.Errorf("setting skip_resize_disk to true when disk_image is false should have error")
	}
}

func TestBuilderPrepare_FloppyFiles(t *testing.T) {
	var c Config
	config := testConfig()

	delete(config, "floppy_files")
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if len(c.FloppyFiles) != 0 {
		t.Fatalf("bad: %#v", c.FloppyFiles)
	}

	floppies_path := "../../packer-plugin-sdk/test-fixtures/floppies"
	config["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	if !reflect.DeepEqual(c.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", c.FloppyFiles)
	}
}

func TestBuilderPrepare_InvalidFloppies(t *testing.T) {
	var c Config
	config := testConfig()
	config["floppy_files"] = []string{"nonexistent.bat", "nonexistent.ps1"}
	c = Config{}
	_, errs := c.Prepare(config)
	if errs == nil {
		t.Fatalf("Nonexistent floppies should trigger multierror")
	}

	if len(errs.(*packersdk.MultiError).Errors) != 2 {
		t.Fatalf("Multierror should work and report 2 errors")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var c Config
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_OutputDir(t *testing.T) {
	var c Config
	config := testConfig()

	// Test with existing dir
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	config["output_directory"] = dir
	c = Config{}
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["output_directory"] = "i-hope-i-dont-exist"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ShutdownTimeout(t *testing.T) {
	var c Config
	config := testConfig()

	// Test with a bad value
	config["shutdown_timeout"] = "this is not good"
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["shutdown_timeout"] = "5s"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHHostPort(t *testing.T) {
	var c Config
	config := testConfig()

	// Bad
	config["host_port_min"] = 1000
	config["host_port_max"] = 500
	c = Config{}
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["host_port_min"] = -500
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["host_port_min"] = 500
	config["host_port_max"] = 1000
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHPrivateKey(t *testing.T) {
	var c Config
	config := testConfig()

	config["ssh_private_key_file"] = ""
	c = Config{}
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["ssh_private_key_file"] = "/i/dont/exist"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad contents
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err := tf.Write([]byte("HELLO!")); err != nil {
		t.Fatalf("err: %s", err)
	}

	config["ssh_private_key_file"] = tf.Name()
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good contents
	if _, err := tf.Seek(0, 0); err != nil {
		t.Fatalf("errorf getting key")
	}
	if err := tf.Truncate(0); err != nil {
		t.Fatalf("errorf getting key")
	}
	if _, err := tf.Write([]byte(testPem)); err != nil {
		t.Fatalf("errorf getting key")
	}
	config["ssh_private_key_file"] = tf.Name()
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestBuilderPrepare_SSHWaitTimeout(t *testing.T) {
	var c Config
	config := testConfig()

	// Test a default boot_wait
	delete(config, "ssh_timeout")
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a bad value
	config["ssh_timeout"] = "this is not good"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["ssh_timeout"] = "5s"
	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_QemuArgs(t *testing.T) {
	var c Config
	config := testConfig()

	// Test with empty
	delete(config, "qemuargs")
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(c.QemuArgs, [][]string{}) {
		t.Fatalf("bad: %#v", c.QemuArgs)
	}

	// Test with a good one
	config["qemuargs"] = [][]interface{}{
		{"foo", "bar", "baz"},
	}

	c = Config{}
	warns, err = c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := [][]string{
		{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(c.QemuArgs, expected) {
		t.Fatalf("bad: %#v", c.QemuArgs)
	}
}

func TestBuilderPrepare_VNCPassword(t *testing.T) {
	var c Config
	config := testConfig()
	config["vnc_use_password"] = true
	config["output_directory"] = "not-a-real-directory"

	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := filepath.Join("not-a-real-directory", "packer-foo.monitor")
	if !reflect.DeepEqual(c.QMPSocketPath, expected) {
		t.Fatalf("Bad QMP socket Path: %s", c.QMPSocketPath)
	}
}

func TestCommConfigPrepare_BackwardsCompatibility(t *testing.T) {
	var c Config
	config := testConfig()
	hostPortMin := 1234
	hostPortMax := 4321
	sshTimeout := 2 * time.Minute

	config["ssh_wait_timeout"] = sshTimeout
	config["ssh_host_port_min"] = hostPortMin
	config["ssh_host_port_max"] = hostPortMax

	warns, err := c.Prepare(config)
	if len(warns) == 0 {
		t.Fatalf("should have deprecation warn")
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if c.CommConfig.Comm.SSHTimeout != sshTimeout {
		t.Fatalf("SSHTimeout should be %s for backwards compatibility, but it was %s", sshTimeout.String(), c.CommConfig.Comm.SSHTimeout.String())
	}

	if c.CommConfig.HostPortMin != hostPortMin {
		t.Fatalf("HostPortMin should be %d for backwards compatibility, but it was %d", hostPortMin, c.CommConfig.HostPortMin)
	}

	if c.CommConfig.HostPortMax != hostPortMax {
		t.Fatalf("HostPortMax should be %d for backwards compatibility, but it was %d", hostPortMax, c.CommConfig.HostPortMax)
	}
}

func TestBuilderPrepare_LoadQemuImgArgs(t *testing.T) {
	var c Config
	config := testConfig()
	config["qemu_img_args"] = map[string][]string{
		"convert": []string{"-o", "preallocation=full"},
		"resize":  []string{"-foo", "bar"},
		"create":  []string{"-baz", "bang"},
	}
	warns, err := c.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	assert.Equal(t, []string{"-o", "preallocation=full"},
		c.QemuImgArgs.Convert, "Convert args not loaded properly")
	assert.Equal(t, []string{"-foo", "bar"},
		c.QemuImgArgs.Resize, "Resize args not loaded properly")
	assert.Equal(t, []string{"-baz", "bang"},
		c.QemuImgArgs.Create, "Create args not loaded properly")
}

package virtualbox

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
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
		"iso_checksum":      "foo",
		"iso_checksum_type": "md5",
		"iso_url":           "http://www.google.com/",
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

func TestBuilderPrepare_Defaults(t *testing.T) {
	var b Builder
	config := testConfig()
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestAdditionsMode != GuestAdditionsModeUpload {
		t.Errorf("bad guest additions mode: %s", b.config.GuestAdditionsMode)
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

	if b.config.Format != "ovf" {
		t.Errorf("bad format: %s", b.config.Format)
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
	b = Builder{}
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

func TestBuilderPrepare_FloppyFiles(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "floppy_files")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	if len(b.config.FloppyFiles) != 0 {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}

	config["floppy_files"] = []string{"foo", "bar"}
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual(b.config.FloppyFiles, expected) {
		t.Fatalf("bad: %#v", b.config.FloppyFiles)
	}
}

func TestBuilderPrepare_GuestAdditionsMode(t *testing.T) {
	var b Builder
	config := testConfig()

	// test default mode
	delete(config, "guest_additions_mode")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("bad err: %s", err)
	}

	// Test another mode
	config["guest_additions_mode"] = "attach"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.GuestAdditionsMode != GuestAdditionsModeAttach {
		t.Fatalf("bad: %s", b.config.GuestAdditionsMode)
	}

	// Test bad mode
	config["guest_additions_mode"] = "teleport"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should error")
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

	config["guest_additions_url"] = "http://www.packer.io"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_HardDriveInterface(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "hard_drive_interface")
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.HardDriveInterface != "ide" {
		t.Fatalf("bad: %s", b.config.HardDriveInterface)
	}

	// Test with a bad
	config["hard_drive_interface"] = "fake"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good
	config["hard_drive_interface"] = "sata"
	b = Builder{}
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
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["http_port_min"] = 500
	config["http_port_max"] = 1000
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_Format(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["format"] = "illegal value"
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["format"] = "ova"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Good
	config["format"] = "ovf"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
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

func TestBuilderPrepare_ISOChecksum(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test bad
	config["iso_checksum"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum"] = "FOo"
	b = Builder{}
	err = b.Prepare(config)
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
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	config["iso_checksum_type"] = "mD5"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ISOChecksumType != "md5" {
		t.Fatalf("should've lowercased: %s", b.config.ISOChecksumType)
	}

	// Test unknown
	config["iso_checksum_type"] = "fake"
	b = Builder{}
	err = b.Prepare(config)
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
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test iso_url set
	config["iso_url"] = "http://www.packer.io"
	b = Builder{}
	err = b.Prepare(config)
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
	err = b.Prepare(config)
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
	err = b.Prepare(config)
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
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["output_directory"] = "i-hope-i-dont-exist"
	b = Builder{}
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
	b = Builder{}
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
	b = Builder{}
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["ssh_host_port_min"] = -500
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["ssh_host_port_min"] = 500
	config["ssh_host_port_max"] = 1000
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_sshKeyPath(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ssh_key_path"] = ""
	b = Builder{}
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["ssh_key_path"] = "/i/dont/exist"
	b = Builder{}
	err = b.Prepare(config)
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

	config["ssh_key_path"] = tf.Name()
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good contents
	tf.Seek(0, 0)
	tf.Truncate(0)
	tf.Write([]byte(testPem))
	config["ssh_key_path"] = tf.Name()
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestBuilderPrepare_SSHUser(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ssh_username"] = ""
	b = Builder{}
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["ssh_username"] = "exists"
	b = Builder{}
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
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["ssh_wait_timeout"] = "5s"
	b = Builder{}
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

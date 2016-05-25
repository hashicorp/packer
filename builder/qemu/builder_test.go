package qemu

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
		"iso_checksum":            "foo",
		"iso_checksum_type":       "md5",
		"iso_url":                 "http://www.google.com/",
		"ssh_username":            "foo",
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

	if b.config.OutputDir != "output-foo" {
		t.Errorf("bad output dir: %s", b.config.OutputDir)
	}

	if b.config.SSHHostPortMin != 2222 {
		t.Errorf("bad min ssh host port: %d", b.config.SSHHostPortMin)
	}

	if b.config.SSHHostPortMax != 4444 {
		t.Errorf("bad max ssh host port: %d", b.config.SSHHostPortMax)
	}

	if b.config.Comm.SSHPort != 22 {
		t.Errorf("bad ssh port: %d", b.config.Comm.SSHPort)
	}

	if b.config.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", b.config.VMName)
	}

	if b.config.Format != "qcow2" {
		t.Errorf("bad format: %s", b.config.Format)
	}
}

func TestBuilderPrepare_BootWait(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "boot_wait")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.RawBootWait != "10s" {
		t.Fatalf("bad value: %s", b.config.RawBootWait)
	}

	// Test with a bad boot_wait
	config["boot_wait"] = "this is not good"
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["boot_wait"] = "5s"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_VNCBindAddress(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "vnc_bind_address")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.VNCBindAddress != "127.0.0.1" {
		t.Fatalf("bad value: %s", b.config.VNCBindAddress)
	}
}

func TestBuilderPrepare_DiskCompaction(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["skip_compaction"] = false
	config["disk_compression"] = true
	config["format"] = "img"
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}
	if b.config.SkipCompaction != true {
		t.Fatalf("SkipCompaction should be true")
	}
	if b.config.DiskCompression != false {
		t.Fatalf("DiskCompression should be false")
	}

	// Good
	config["skip_compaction"] = false
	config["disk_compression"] = true
	config["format"] = "qcow2"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if b.config.SkipCompaction != false {
		t.Fatalf("SkipCompaction should be false")
	}
	if b.config.DiskCompression != true {
		t.Fatalf("DiskCompression should be true")
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

func TestBuilderPrepare_Format(t *testing.T) {
	var b Builder
	config := testConfig()

	// Bad
	config["format"] = "illegal value"
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["format"] = "qcow2"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Good
	config["format"] = "raw"
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

func TestBuilderPrepare_ShutdownTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with a bad value
	config["shutdown_timeout"] = "this is not good"
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["shutdown_timeout"] = "5s"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
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
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Bad
	config["ssh_host_port_min"] = -500
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Good
	config["ssh_host_port_min"] = 500
	config["ssh_host_port_max"] = 1000
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHPrivateKey(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ssh_private_key_file"] = ""
	b = Builder{}
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["ssh_private_key_file"] = "/i/dont/exist"
	b = Builder{}
	warns, err = b.Prepare(config)
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
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good contents
	tf.Seek(0, 0)
	tf.Truncate(0)
	tf.Write([]byte(testPem))
	config["ssh_private_key_file"] = tf.Name()
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestBuilderPrepare_SSHUser(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ssh_username"] = ""
	b = Builder{}
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["ssh_username"] = "exists"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_SSHWaitTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test a default boot_wait
	delete(config, "ssh_wait_timeout")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a bad value
	config["ssh_wait_timeout"] = "this is not good"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["ssh_wait_timeout"] = "5s"
	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_QemuArgs(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with empty
	delete(config, "qemuargs")
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(b.config.QemuArgs, [][]string{}) {
		t.Fatalf("bad: %#v", b.config.QemuArgs)
	}

	// Test with a good one
	config["qemuargs"] = [][]interface{}{
		[]interface{}{"foo", "bar", "baz"},
	}

	b = Builder{}
	warns, err = b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := [][]string{
		[]string{"foo", "bar", "baz"},
	}

	if !reflect.DeepEqual(b.config.QemuArgs, expected) {
		t.Fatalf("bad: %#v", b.config.QemuArgs)
	}
}

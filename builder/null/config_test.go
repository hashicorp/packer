// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package null

import (
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"ssh_host":     "foo",
		"ssh_username": "bar",
		"ssh_password": "baz",
	}
}

func testConfigStruct(t *testing.T) *Config {
	var c Config
	warns, errs := c.Prepare(testConfig())
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return &c
}

func testConfigErr(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestConfigPrepare_port(t *testing.T) {
	raw := testConfig()

	// default port should be 22
	delete(raw, "port")
	var c Config
	warns, errs := c.Prepare(raw)
	if c.CommConfig.SSHPort != 22 {
		t.Fatalf("bad: port should default to 22, not %d", c.CommConfig.SSHPort)
	}
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_host(t *testing.T) {
	raw := testConfig()

	// No host
	delete(raw, "ssh_host")
	var c Config
	warns, errs := c.Prepare(raw)
	testConfigErr(t, warns, errs)

	// Good host
	raw["ssh_host"] = "good"
	warns, errs = c.Prepare(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_sshCredential(t *testing.T) {
	raw := testConfig()

	// no ssh_password and no ssh_private_key_file
	delete(raw, "ssh_password")
	delete(raw, "ssh_private_key_file")
	warns, errs := (&Config{}).Prepare(raw)
	testConfigErr(t, warns, errs)

	// only ssh_password
	raw["ssh_password"] = "good"
	warns, errs = (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)

	// only ssh_private_key_file
	testFile := communicator.TestPEM(t)
	defer os.Remove(testFile)
	raw["ssh_private_key_file"] = testFile
	delete(raw, "ssh_password")
	warns, errs = (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)

	// both ssh_password and ssh_private_key_file set
	raw["ssh_password"] = "bad"
	warns, errs = (&Config{}).Prepare(raw)
	testConfigErr(t, warns, errs)
}

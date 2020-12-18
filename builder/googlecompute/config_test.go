package googlecompute

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
)

func TestConfigPrepare(t *testing.T) {
	cases := []struct {
		Key   string
		Value interface{}
		Err   bool
	}{
		{
			"unknown_key",
			"bad",
			true,
		},

		{
			"private_key_file",
			"/tmp/i/should/not/exist",
			true,
		},

		{
			"project_id",
			nil,
			true,
		},
		{
			"project_id",
			"foo",
			false,
		},

		{
			"source_image",
			nil,
			true,
		},
		{
			"source_image",
			"foo",
			false,
		},

		{
			"source_image_family",
			nil,
			false,
		},
		{
			"source_image_family",
			"foo",
			false,
		},

		{
			"zone",
			nil,
			true,
		},
		{
			"zone",
			"foo",
			false,
		},

		{
			"ssh_timeout",
			"SO BAD",
			true,
		},
		{
			"ssh_timeout",
			"5s",
			false,
		},

		{
			"wait_to_add_ssh_keys",
			"SO BAD",
			true,
		},
		{
			"wait_to_add_ssh_keys",
			"5s",
			false,
		},

		{
			"state_timeout",
			"SO BAD",
			true,
		},
		{
			"state_timeout",
			"5s",
			false,
		},
		{
			"use_internal_ip",
			nil,
			false,
		},
		{
			"use_internal_ip",
			false,
			false,
		},
		{
			"use_internal_ip",
			"SO VERY BAD",
			true,
		},
		{
			"on_host_maintenance",
			nil,
			false,
		},
		{
			"on_host_maintenance",
			"TERMINATE",
			false,
		},
		{
			"on_host_maintenance",
			"SO VERY BAD",
			true,
		},
		{
			"preemptible",
			nil,
			false,
		},
		{
			"preemptible",
			false,
			false,
		},
		{
			"preemptible",
			"SO VERY BAD",
			true,
		},
		{
			"image_family",
			nil,
			false,
		},
		{
			"image_family",
			"",
			false,
		},
		{
			"image_family",
			"foo-bar",
			false,
		},
		{
			"image_family",
			"foo bar",
			true,
		},
		{
			// underscore is not allowed
			"image_name",
			"foo_bar",
			true,
		},
		{
			// too long
			"image_name",
			"foobar123xyz_abc-456-one-two_three_five_nine_seventeen_eleventy-seven",
			true,
		},
		{
			// starts with non-alphabetic character
			"image_name",
			"1boohoo",
			true,
		},
		{
			"image_encryption_key",
			map[string]string{"kmsKeyName": "foo"},
			false,
		},
		{
			"image_encryption_key",
			map[string]string{"No such key": "foo"},
			true,
		},
		{
			"image_encryption_key",
			map[string]string{"kmsKeyName": "foo", "RawKey": "foo"},
			false,
		},
		{
			"scopes",
			[]string{},
			false,
		},
		{
			"scopes",
			[]string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/devstorage.full_control", "https://www.googleapis.com/auth/sqlservice.admin"},
			false,
		},
		{
			"scopes",
			[]string{"https://www.googleapis.com/auth/cloud-platform"},
			false,
		},

		{
			"disable_default_service_account",
			"",
			false,
		},
		{
			"disable_default_service_account",
			nil,
			false,
		},
		{
			"disable_default_service_account",
			false,
			false,
		},
		{
			"disable_default_service_account",
			true,
			false,
		},
		{
			"disable_default_service_account",
			"NOT A BOOL",
			true,
		},
	}

	for _, tc := range cases {
		raw, tempfile := testConfig(t)
		defer os.Remove(tempfile)

		if tc.Value == nil {
			delete(raw, tc.Key)
		} else {
			raw[tc.Key] = tc.Value
		}

		var c Config
		warns, errs := c.Prepare(raw)

		if tc.Err {
			testConfigErr(t, warns, errs, tc.Key)
		} else {
			testConfigOk(t, warns, errs)
		}
	}
}

func TestConfigPrepareAccelerator(t *testing.T) {
	cases := []struct {
		Keys   []string
		Values []interface{}
		Err    bool
	}{
		{
			[]string{"accelerator_count", "on_host_maintenance", "accelerator_type"},
			[]interface{}{1, "MIGRATE", "something_valid"},
			true,
		},
		{
			[]string{"accelerator_count", "on_host_maintenance", "accelerator_type"},
			[]interface{}{1, "TERMINATE", "something_valid"},
			false,
		},
		{
			[]string{"accelerator_count", "on_host_maintenance", "accelerator_type"},
			[]interface{}{1, "TERMINATE", nil},
			true,
		},
		{
			[]string{"accelerator_count", "on_host_maintenance", "accelerator_type"},
			[]interface{}{1, "TERMINATE", ""},
			true,
		},
		{
			[]string{"accelerator_count", "on_host_maintenance", "accelerator_type"},
			[]interface{}{1, "TERMINATE", "something_valid"},
			false,
		},
	}

	for _, tc := range cases {
		raw, tempfile := testConfig(t)
		defer os.Remove(tempfile)

		errStr := ""
		for k := range tc.Keys {

			// Create the string for error reporting
			// convert value to string if it can be converted
			errStr += fmt.Sprintf("%s:%v, ", tc.Keys[k], tc.Values[k])
			if tc.Values[k] == nil {
				delete(raw, tc.Keys[k])
			} else {
				raw[tc.Keys[k]] = tc.Values[k]
			}
		}

		var c Config
		warns, errs := c.Prepare(raw)

		if tc.Err {
			testConfigErr(t, warns, errs, strings.TrimRight(errStr, ", "))
		} else {
			testConfigOk(t, warns, errs)
		}
	}
}

func TestConfigPrepareServiceAccount(t *testing.T) {
	cases := []struct {
		Keys   []string
		Values []interface{}
		Err    bool
	}{
		{
			[]string{"disable_default_service_account", "service_account_email"},
			[]interface{}{true, "service@account.email.com"},
			true,
		},
		{
			[]string{"disable_default_service_account", "service_account_email"},
			[]interface{}{false, "service@account.email.com"},
			false,
		},
		{
			[]string{"disable_default_service_account", "service_account_email"},
			[]interface{}{true, ""},
			false,
		},
	}

	for _, tc := range cases {
		raw, tempfile := testConfig(t)
		defer os.Remove(tempfile)

		errStr := ""
		for k := range tc.Keys {

			// Create the string for error reporting
			// convert value to string if it can be converted
			errStr += fmt.Sprintf("%s:%v, ", tc.Keys[k], tc.Values[k])
			if tc.Values[k] == nil {
				delete(raw, tc.Keys[k])
			} else {
				raw[tc.Keys[k]] = tc.Values[k]
			}
		}

		var c Config
		warns, errs := c.Prepare(raw)

		if tc.Err {
			testConfigErr(t, warns, errs, strings.TrimRight(errStr, ", "))
		} else {
			testConfigOk(t, warns, errs)
		}
	}
}

func TestConfigPrepareStartupScriptFile(t *testing.T) {
	config := map[string]interface{}{
		"project_id":          "project",
		"source_image":        "foo",
		"ssh_username":        "packer",
		"startup_script_file": "no-such-file",
		"zone":                "us-central1-a",
	}

	var c Config
	_, errs := c.Prepare(config)

	if errs == nil || !strings.Contains(errs.Error(), "startup_script_file") {
		t.Fatalf("should error: startup_script_file")
	}
}

func TestConfigPrepareIAP_SSH(t *testing.T) {
	config := map[string]interface{}{
		"project_id":   "project",
		"source_image": "foo",
		"ssh_username": "packer",
		"zone":         "us-central1-a",
		"communicator": "ssh",
		"use_iap":      true,
	}

	var c Config
	_, err := c.Prepare(config)
	if err != nil {
		t.Fatalf("Shouldn't have errors. Err = %s", err)
	}
	if c.Comm.SSHHost != "localhost" {
		t.Fatalf("Should have set SSHHost")
	}

	testIAPScript(t, &c)
}

func TestConfigPrepareIAP_WinRM(t *testing.T) {
	config := map[string]interface{}{
		"project_id":     "project",
		"source_image":   "foo",
		"winrm_username": "packer",
		"zone":           "us-central1-a",
		"communicator":   "winrm",
		"use_iap":        true,
	}

	var c Config
	_, err := c.Prepare(config)
	if err != nil {
		t.Fatalf("Shouldn't have errors. Err = %s", err)
	}
	if c.Comm.WinRMHost != "localhost" {
		t.Fatalf("Should have set WinRMHost")
	}

	testIAPScript(t, &c)
}

func TestConfigPrepareIAP_failures(t *testing.T) {
	config := map[string]interface{}{
		"project_id":     "project",
		"source_image":   "foo",
		"winrm_username": "packer",
		"zone":           "us-central1-a",
		"communicator":   "none",
		"iap_hashbang":   "/bin/bash",
		"iap_ext":        ".ps1",
		"use_iap":        true,
	}

	var c Config
	_, errs := c.Prepare(config)
	if errs == nil {
		t.Fatalf("Should have errored because we're using none.")
	}
	if c.IAPHashBang != "/bin/bash" {
		t.Fatalf("IAP hashbang defaulted even though set.")
	}
	if c.IAPExt != ".ps1" {
		t.Fatalf("IAP tempfile defaulted even though set.")
	}
}

func TestConfigDefaults(t *testing.T) {
	cases := []struct {
		Read  func(c *Config) interface{}
		Value interface{}
	}{
		{
			func(c *Config) interface{} { return c.Comm.Type },
			"ssh",
		},

		{
			func(c *Config) interface{} { return c.Comm.SSHPort },
			22,
		},
	}

	for _, tc := range cases {
		raw, tempfile := testConfig(t)
		defer os.Remove(tempfile)

		var c Config
		warns, errs := c.Prepare(raw)
		testConfigOk(t, warns, errs)

		actual := tc.Read(&c)
		if actual != tc.Value {
			t.Fatalf("bad: %#v", actual)
		}
	}
}

func TestImageName(t *testing.T) {
	raw, tempfile := testConfig(t)
	defer os.Remove(tempfile)

	var c Config
	c.Prepare(raw)
	if !strings.HasPrefix(c.ImageName, "packer-") {
		t.Fatalf("ImageName should have 'packer-' prefix, found %s", c.ImageName)
	}
	if strings.Contains(c.ImageName, "{{timestamp}}") {
		t.Errorf("ImageName should be interpolated; found %s", c.ImageName)
	}
}

func TestRegion(t *testing.T) {
	raw, tempfile := testConfig(t)
	defer os.Remove(tempfile)

	var c Config
	c.Prepare(raw)
	if c.Region != "us-east1" {
		t.Fatalf("Region should be 'us-east1' given Zone of 'us-east1-a', but is %s", c.Region)
	}
}

func TestApplyIAPTunnel_SSH(t *testing.T) {
	c := &communicator.Config{
		Type: "ssh",
		SSH: communicator.SSH{
			SSHHost: "example",
			SSHPort: 1234,
		},
	}

	err := ApplyIAPTunnel(c, 8447)
	if err != nil {
		t.Fatalf("Shouldn't have errors")
	}
	if c.SSHPort != 8447 {
		t.Fatalf("Should have set SSHPort")
	}
}

func TestApplyIAPTunnel_WinRM(t *testing.T) {
	c := &communicator.Config{
		Type: "winrm",
		WinRM: communicator.WinRM{
			WinRMHost: "example",
			WinRMPort: 1234,
		},
	}

	err := ApplyIAPTunnel(c, 8447)
	if err != nil {
		t.Fatalf("Shouldn't have errors")
	}
	if c.WinRMPort != 8447 {
		t.Fatalf("Should have set WinRMPort")
	}
}

func TestApplyIAPTunnel_none(t *testing.T) {
	c := &communicator.Config{
		Type: "none",
	}

	err := ApplyIAPTunnel(c, 8447)
	if err == nil {
		t.Fatalf("Should have errors, none is not supported")
	}
}

// Helper stuff below

func testConfig(t *testing.T) (config map[string]interface{}, tempAccountFile string) {
	tempAccountFile = testAccountFile(t)

	config = map[string]interface{}{
		"account_file": tempAccountFile,
		"project_id":   "hashicorp",
		"source_image": "foo",
		"ssh_username": "root",
		"image_family": "bar",
		"image_labels": map[string]string{
			"label-1": "value-1",
			"label-2": "value-2",
		},
		"image_licenses": []string{
			"test-license",
		},
		"image_storage_locations": []string{
			"us-east1",
		},
		"metadata_files": map[string]string{},
		"zone":           "us-east1-a",
	}

	return config, tempAccountFile
}

func testConfigStruct(t *testing.T) *Config {
	raw, tempfile := testConfig(t)
	defer os.Remove(tempfile)

	var c Config
	warns, errs := c.Prepare(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return &c
}

func testConfigErr(t *testing.T, warns []string, err error, extra string) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should error: %s", extra)
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

func testAccountFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	if _, err := tf.Write([]byte(testAccountContent)); err != nil {
		t.Fatalf("err: %s", err)
	}

	return tf.Name()
}

func testIAPScript(t *testing.T, c *Config) {
	if runtime.GOOS == "windows" {
		if c.IAPExt != ".cmd" {
			t.Fatalf("IAP tempfile extension didn't default correctly to .cmd")
		}
		if c.IAPHashBang != "" {
			t.Fatalf("IAP hashbang didn't default correctly to nothing.")
		}
	} else {
		if c.IAPExt != "" {
			t.Fatalf("IAP tempfile extension should default to empty on unix mahcines")
		}
		if c.IAPHashBang != "/bin/sh" {
			t.Fatalf("IAP hashbang didn't default correctly to /bin/sh.")
		}
	}
}

const testMetadataFileContent = `testMetadata`

func testMetadataFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()
	if _, err := tf.Write([]byte(testMetadataFileContent)); err != nil {
		t.Fatalf("err: %s", err)
	}

	return tf.Name()
}

// This is just some dummy data that doesn't actually work
const testAccountContent = `{
  "type": "service_account",
  "project_id": "test-project-123456789",
  "private_key_id": "bananaphone",
  "private_key": "-----BEGIN PRIVATE KEY-----\nring_ring_ring_ring_ring_ring_ring_BANANAPHONE\n-----END PRIVATE KEY-----\n",
  "client_email": "raffi-compute@developer.gserviceaccount.com",
  "client_id": "1234567890",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/12345-compute%40developer.gserviceaccount.com"
}`

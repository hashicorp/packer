package googlecompute

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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

		_, warns, errs := NewConfig(raw)

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

		_, warns, errs := NewConfig(raw)

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

		_, warns, errs := NewConfig(raw)

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

	_, _, errs := NewConfig(config)

	if errs == nil || !strings.Contains(errs.Error(), "startup_script_file") {
		t.Fatalf("should error: startup_script_file")
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

		c, warns, errs := NewConfig(raw)
		testConfigOk(t, warns, errs)

		actual := tc.Read(c)
		if actual != tc.Value {
			t.Fatalf("bad: %#v", actual)
		}
	}
}

func TestImageName(t *testing.T) {
	raw, tempfile := testConfig(t)
	defer os.Remove(tempfile)

	c, _, _ := NewConfig(raw)
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

	c, _, _ := NewConfig(raw)
	if c.Region != "us-east1" {
		t.Fatalf("Region should be 'us-east1' given Zone of 'us-east1-a', but is %s", c.Region)
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
		"zone": "us-east1-a",
	}

	return config, tempAccountFile
}

func testConfigStruct(t *testing.T) *Config {
	raw, tempfile := testConfig(t)
	defer os.Remove(tempfile)

	c, warns, errs := NewConfig(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return c
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

// This is just some dummy data that doesn't actually work (it was revoked
// a long time ago).
const testAccountContent = `{}`

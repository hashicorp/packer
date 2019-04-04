package yandex

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const TestServiceAccountKeyFile = "./test_data/fake-sa-key.json"

func TestConfigPrepare(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	tf.Close()

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
			"service_account_key_file",
			"/tmp/i/should/not/exist",
			true,
		},
		{
			"service_account_key_file",
			tf.Name(),
			true,
		},
		{
			"service_account_key_file",
			TestServiceAccountKeyFile,
			false,
		},

		{
			"folder_id",
			nil,
			true,
		},
		{
			"folder_id",
			"foo",
			false,
		},

		{
			"source_image_id",
			nil,
			true,
		},
		{
			"source_image_id",
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
			false,
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
	}

	for _, tc := range cases {
		raw := testConfig(t)

		if tc.Value == nil {
			delete(raw, tc.Key)
		} else {
			raw[tc.Key] = tc.Value
		}

		if tc.Key == "service_account_key_file" {
			delete(raw, "token")
		}

		_, warns, errs := NewConfig(raw)

		if tc.Err {
			testConfigErr(t, warns, errs, tc.Key)
		} else {
			testConfigOk(t, warns, errs)
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	cases := []struct {
		Read  func(c *Config) interface{}
		Value interface{}
	}{
		{
			func(c *Config) interface{} { return c.Communicator.Type },
			"ssh",
		},

		{
			func(c *Config) interface{} { return c.Communicator.SSHPort },
			22,
		},
	}

	for _, tc := range cases {
		raw := testConfig(t)

		c, warns, errs := NewConfig(raw)
		testConfigOk(t, warns, errs)

		actual := tc.Read(c)
		if actual != tc.Value {
			t.Fatalf("bad: %#v", actual)
		}
	}
}

func TestImageName(t *testing.T) {
	raw := testConfig(t)

	c, _, _ := NewConfig(raw)
	if !strings.HasPrefix(c.ImageName, "packer-") {
		t.Fatalf("ImageName should have 'packer-' prefix, found %s", c.ImageName)
	}
	if strings.Contains(c.ImageName, "{{timestamp}}") {
		t.Errorf("ImageName should be interpolated; found %s", c.ImageName)
	}
}

func TestZone(t *testing.T) {
	raw := testConfig(t)

	c, _, _ := NewConfig(raw)
	if c.Zone != "ru-central1-a" {
		t.Fatalf("Zone should be 'ru-central1-a' given, but is '%s'", c.Zone)
	}
}

// Helper stuff below

func testConfig(t *testing.T) (config map[string]interface{}) {

	config = map[string]interface{}{
		"token":           "test_token",
		"folder_id":       "hashicorp",
		"source_image_id": "foo",
		"ssh_username":    "root",
		"image_family":    "bar",
		"image_product_ids": []string{
			"test-license",
		},
		"zone": "ru-central1-a",
	}

	return config
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

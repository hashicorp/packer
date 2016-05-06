package googlecompute

import (
	"io/ioutil"
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
	}

	for _, tc := range cases {
		raw := testConfig(t)

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
	c, _, _ := NewConfig(testConfig(t))
	if !strings.HasPrefix(c.ImageName, "packer-") {
		t.Fatalf("ImageName should have 'packer-' prefix, found %s", c.ImageName)
	}
	if strings.Contains(c.ImageName, "{{timestamp}}") {
		t.Errorf("ImageName should be interpolated; found %s", c.ImageName)
	}
}

func TestRegion(t *testing.T) {
	c, _, _ := NewConfig(testConfig(t))
	if c.Region != "us-east1" {
		t.Fatalf("Region should be 'us-east1' given Zone of 'us-east1-a', but is %s", c.Region)
	}
}

// Helper stuff below

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"account_file": testAccountFile(t),
		"project_id":   "hashicorp",
		"source_image": "foo",
		"image_family": "bar",
		"zone":         "us-east1-a",
	}
}

func testConfigStruct(t *testing.T) *Config {
	c, warns, errs := NewConfig(testConfig(t))
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

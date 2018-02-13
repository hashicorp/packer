package oci

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
)

func testConfig(accessConfFile *os.File) map[string]interface{} {
	return map[string]interface{}{
		"availability_domain": "aaaa:PHX-AD-3",
		"access_cfg_file":     accessConfFile.Name(),

		// Image
		"base_image_ocid": "ocd1...",
		"shape":           "VM.Standard1.1",
		"image_name":      "HelloWorld",

		// Networking
		"subnet_ocid": "ocd1...",

		// Comm
		"ssh_username": "opc",
	}
}

func getField(c *client.Config, field string) string {
	r := reflect.ValueOf(c)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

func TestConfig(t *testing.T) {
	// Shared set-up and defered deletion

	cfg, keyFile, err := client.BaseTestConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())

	cfgFile, err := client.WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfgFile.Name())

	// Temporarily set $HOME to temp directory to bypass default
	// access config loading.

	tmpHome, err := ioutil.TempDir("", "packer_config_test")
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	defer os.Remove(tmpHome)

	home := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", home)

	// Config tests

	t.Run("BaseConfig", func(t *testing.T) {
		raw := testConfig(cfgFile)
		_, errs := NewConfig(raw)

		if errs != nil {
			t.Fatalf("err: %+v", errs)
		}

	})

	t.Run("NoAccessConfig", func(t *testing.T) {
		raw := testConfig(cfgFile)
		delete(raw, "access_cfg_file")

		_, errs := NewConfig(raw)

		s := errs.Error()
		expectedErrors := []string{
			"'user_ocid'", "'tenancy_ocid'", "'fingerprint'",
			"'key_file'",
		}
		for _, expected := range expectedErrors {
			if !strings.Contains(s, expected) {
				t.Errorf("Expected %s to contain '%s'", s, expected)
			}
		}
	})

	t.Run("AccessConfigTemplateOnly", func(t *testing.T) {
		raw := testConfig(cfgFile)
		delete(raw, "access_cfg_file")
		raw["user_ocid"] = "ocid1..."
		raw["tenancy_ocid"] = "ocid1..."
		raw["fingerprint"] = "00:00..."
		raw["key_file"] = keyFile.Name()

		_, errs := NewConfig(raw)

		if errs != nil {
			t.Fatalf("err: %+v", errs)
		}

	})

	t.Run("TenancyReadFromAccessCfgFile", func(t *testing.T) {
		raw := testConfig(cfgFile)
		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("err: %+v", errs)
		}

		expected := "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		if c.AccessCfg.Tenancy != expected {
			t.Errorf("Expected tenancy: %s, got %s.", expected, c.AccessCfg.Tenancy)
		}

	})

	t.Run("RegionNotDefaultedToPHXWhenSetInOCISettings", func(t *testing.T) {
		raw := testConfig(cfgFile)
		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("err: %+v", errs)
		}

		expected := "us-ashburn-1"
		if c.AccessCfg.Region != expected {
			t.Errorf("Expected region: %s, got %s.", expected, c.AccessCfg.Region)
		}

	})

	// Test the correct errors are produced when required template keys are
	// omitted.
	requiredKeys := []string{"availability_domain", "base_image_ocid", "shape", "subnet_ocid"}
	for _, k := range requiredKeys {
		t.Run(k+"_required", func(t *testing.T) {
			raw := testConfig(cfgFile)
			delete(raw, k)

			_, errs := NewConfig(raw)

			if !strings.Contains(errs.Error(), k) {
				t.Errorf("Expected '%s' to contain '%s'", errs.Error(), k)
			}
		})
	}

	t.Run("ImageNameDefaultedIfEmpty", func(t *testing.T) {
		raw := testConfig(cfgFile)
		delete(raw, "image_name")

		c, errs := NewConfig(raw)
		if errs != nil {
			t.Errorf("Unexpected error(s): %s", errs)
		}

		if !strings.Contains(c.ImageName, "packer-") {
			t.Errorf("got default ImageName %q, want image name 'packer-{{timestamp}}'", c.ImageName)
		}
	})

	// Test that AccessCfgFile properties are overridden by their
	// corresponding template keys.
	accessOverrides := map[string]string{
		"user_ocid":    "User",
		"tenancy_ocid": "Tenancy",
		"region":       "Region",
		"fingerprint":  "Fingerprint",
	}
	for k, v := range accessOverrides {
		t.Run("AccessCfg."+v+"Overridden", func(t *testing.T) {
			expected := "override"

			raw := testConfig(cfgFile)
			raw[k] = expected

			c, errs := NewConfig(raw)
			if errs != nil {
				t.Fatalf("err: %+v", errs)
			}

			accessVal := getField(c.AccessCfg, v)
			if accessVal != expected {
				t.Errorf("Expected AccessCfg.%s: %s, got %s", v, expected, accessVal)
			}
		})
	}
}

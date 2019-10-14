package oci

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-ini/ini"
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
		"ssh_username":   "opc",
		"use_private_ip": false,
		"metadata": map[string]string{
			"key": "value",
		},
		"defined_tags": map[string]map[string]interface{}{
			"namespace": {"key": "value"},
		},
	}
}

func TestConfig(t *testing.T) {
	// Shared set-up and deferred deletion

	cfg, keyFile, err := baseTestConfigWithTmpKeyFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())

	cfgFile, err := writeTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfgFile.Name())

	// Temporarily set $HOME to temp directory to bypass default
	// access config loading.

	tmpHome, err := ioutil.TempDir("", "packer_config_test")
	if err != nil {
		t.Fatalf("Unexpected error when creating temporary directory: %+v", err)
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
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}
	})

	t.Run("NoAccessConfig", func(t *testing.T) {
		raw := testConfig(cfgFile)
		delete(raw, "access_cfg_file")

		_, errs := NewConfig(raw)

		expectedErrors := []string{
			"'user_ocid'", "'tenancy_ocid'", "'fingerprint'", "'key_file'",
		}

		s := errs.Error()
		for _, expected := range expectedErrors {
			if !strings.Contains(s, expected) {
				t.Errorf("Expected %q to contain '%s'", s, expected)
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
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		tenancy, err := c.configProvider.TenancyOCID()
		if err != nil {
			t.Fatalf("Unexpected error getting tenancy ocid: %v", err)
		}

		expected := "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		if tenancy != expected {
			t.Errorf("Expected tenancy: %s, got %s.", expected, tenancy)
		}

	})

	t.Run("RegionNotDefaultedToPHXWhenSetInOCISettings", func(t *testing.T) {
		raw := testConfig(cfgFile)
		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		region, err := c.configProvider.Region()
		if err != nil {
			t.Fatalf("Unexpected error getting region: %v", err)
		}

		expected := "us-ashburn-1"
		if region != expected {
			t.Errorf("Expected region: %s, got %s.", expected, region)
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
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		if !strings.Contains(c.ImageName, "packer-") {
			t.Errorf("got default ImageName %q, want image name 'packer-{{timestamp}}'", c.ImageName)
		}
	})

	t.Run("user_ocid_overridden", func(t *testing.T) {
		expected := "override"
		raw := testConfig(cfgFile)
		raw["user_ocid"] = expected

		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		user, _ := c.configProvider.UserOCID()
		if user != expected {
			t.Errorf("Expected ConfigProvider.UserOCID: %s, got %s", expected, user)
		}
	})

	t.Run("tenancy_ocid_overidden", func(t *testing.T) {
		expected := "override"
		raw := testConfig(cfgFile)
		raw["tenancy_ocid"] = expected

		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		tenancy, _ := c.configProvider.TenancyOCID()
		if tenancy != expected {
			t.Errorf("Expected ConfigProvider.TenancyOCID: %s, got %s", expected, tenancy)
		}
	})

	t.Run("region_overidden", func(t *testing.T) {
		expected := "override"
		raw := testConfig(cfgFile)
		raw["region"] = expected

		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("Unexpected error in configuration %+v", errs)
		}

		region, _ := c.configProvider.Region()
		if region != expected {
			t.Errorf("Expected ConfigProvider.Region: %s, got %s", expected, region)
		}
	})

	t.Run("fingerprint_overidden", func(t *testing.T) {
		expected := "override"
		raw := testConfig(cfgFile)
		raw["fingerprint"] = expected

		c, errs := NewConfig(raw)
		if errs != nil {
			t.Fatalf("Unexpected error in configuration: %+v", errs)
		}

		fingerprint, _ := c.configProvider.KeyFingerprint()
		if fingerprint != expected {
			t.Errorf("Expected ConfigProvider.KeyFingerprint: %s, got %s", expected, fingerprint)
		}
	})
}

// BaseTestConfig creates the base (DEFAULT) config including a temporary key
// file.
// NOTE: Caller is responsible for removing temporary key file.
func baseTestConfigWithTmpKeyFile() (*ini.File, *os.File, error) {
	keyFile, err := generateRSAKeyFile()
	if err != nil {
		return nil, keyFile, err
	}
	// Build ini
	cfg := ini.Empty()
	section, _ := cfg.NewSection("DEFAULT")
	section.NewKey("region", "us-ashburn-1")
	section.NewKey("tenancy", "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	section.NewKey("user", "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	section.NewKey("fingerprint", "70:04:5z:b3:19:ab:90:75:a4:1f:50:d4:c7:c3:33:20")
	section.NewKey("key_file", keyFile.Name())

	return cfg, keyFile, nil
}

// WriteTestConfig writes a ini.File to a temporary file for use in unit tests.
// NOTE: Caller is responsible for removing temporary file.
func writeTestConfig(cfg *ini.File) (*os.File, error) {
	confFile, err := ioutil.TempFile("", "config_file")
	if err != nil {
		return nil, err
	}

	if _, err := confFile.Write([]byte("[DEFAULT]\n")); err != nil {
		os.Remove(confFile.Name())
		return nil, err
	}

	if _, err := cfg.WriteTo(confFile); err != nil {
		os.Remove(confFile.Name())
		return nil, err
	}
	return confFile, nil
}

// generateRSAKeyFile generates an RSA key file for use in unit tests.
// NOTE: The caller is responsible for deleting the temporary file.
func generateRSAKeyFile() (*os.File, error) {
	// Create temporary file for the key
	f, err := ioutil.TempFile("", "key")
	if err != nil {
		return nil, err
	}

	// Generate key
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Write the key out
	if _, err := f.Write(pem.EncodeToMemory(&privBlk)); err != nil {
		return nil, err
	}

	return f, nil
}

package oci

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

// Config API authentication and target configuration
type Config struct {
	// User OCID e.g. ocid1.user.oc1..aaaaaaaadcshyehbkvxl7arse3lv7z5oknexjgfhnhwidtugsxhlm4247
	User string `ini:"user"`

	// User's Tenancy OCID e.g. ocid1.tenancy.oc1..aaaaaaaagtgvshv6opxzjyzkupkt64ymd32n6kbomadanpcg43d
	Tenancy string `ini:"tenancy"`

	// Bare metal region identifier (e.g. us-phoenix-1)
	Region string `ini:"region"`

	// Hex key fingerprint (e.g. b5:a0:62:57:28:0d:fd:c9:59:16:eb:d4:51:9f:70:e4)
	Fingerprint string `ini:"fingerprint"`

	// Path to OCI config file (e.g. ~/.oci/config)
	KeyFile string `ini:"key_file"`

	// Passphrase used for the key, if it is encrypted.
	PassPhrase string `ini:"pass_phrase"`

	// Private key (loaded via LoadPrivateKey or ParsePrivateKey)
	Key *rsa.PrivateKey

	// Used to override base API URL.
	baseURL string
}

// getBaseURL returns either the specified base URL or builds the appropriate
// URL based on service, region, and API version.
func (c *Config) getBaseURL(service string) string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return fmt.Sprintf(baseURLPattern, service, c.Region, apiVersion)
}

// LoadConfigsFromFile loads all oracle oci configurations from a file
// (generally ~/.oci/config).
func LoadConfigsFromFile(path string) (map[string]*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("Oracle OCI config file is missing: %s", path)
	}

	cfgFile, err := ini.Load(path)
	if err != nil {
		err := fmt.Errorf("Failed to parse config file %s: %s", path, err.Error())
		return nil, err
	}

	configs := make(map[string]*Config)

	// Load DEFAULT section to populate defaults for all other configs
	config, err := loadConfigSection(cfgFile, "DEFAULT", nil)
	if err != nil {
		return nil, err
	}
	configs["DEFAULT"] = config

	// Load other sections.
	for _, sectionName := range cfgFile.SectionStrings() {
		if sectionName == "DEFAULT" {
			continue
		}

		// Map to Config struct with defaults from DEFAULT section.
		config, err := loadConfigSection(cfgFile, sectionName, configs["DEFAULT"])
		if err != nil {
			return nil, err
		}
		configs[sectionName] = config
	}

	return configs, nil
}

// Loads an individual Config object from a ini.Section in the Oracle OCI config
// file.
func loadConfigSection(f *ini.File, sectionName string, config *Config) (*Config, error) {
	if config == nil {
		config = &Config{}
	}

	section, err := f.GetSection(sectionName)
	if err != nil {
		return nil, fmt.Errorf("Config file does not contain a %s section", sectionName)
	}

	if err := section.MapTo(config); err != nil {
		return nil, err
	}

	config.Key, err = LoadPrivateKey(config)
	if err != nil {
		return nil, err
	}

	return config, err
}

// LoadPrivateKey loads private key from disk and parses it.
func LoadPrivateKey(config *Config) (*rsa.PrivateKey, error) {
	// Expand '~' to $HOME
	path, err := homedir.Expand(config.KeyFile)
	if err != nil {
		return nil, err
	}

	// Read and parse API signing key
	keyContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	key, err := ParsePrivateKey(keyContent, []byte(config.PassPhrase))

	return key, err
}

// ParsePrivateKey parses a PEM encoded array of bytes into an rsa.PrivateKey.
// Attempts to decrypt the PEM encoded array of bytes with the given password
// if the PEM encoded byte array is encrypted.
func ParsePrivateKey(content, password []byte) (*rsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode(content)

	if keyBlock == nil {
		return nil, errors.New("could not decode PEM private key")
	}

	var der []byte
	var err error
	if x509.IsEncryptedPEMBlock(keyBlock) {
		if len(password) < 1 {
			return nil, errors.New("encrypted private key but no pass phrase provided")
		}
		der, err = x509.DecryptPEMBlock(keyBlock, password)
		if err != nil {
			return nil, err
		}
	} else {
		der = keyBlock.Bytes
	}

	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(der)
	if err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("Private key is not an RSA private key")
		}
	}
	return nil, fmt.Errorf("Failed to parse private key :%s", err)
}

// BaseTestConfig creates the base (DEFAULT) config including a temporary key
// file.
// NOTE: Caller is responsible for removing temporary key file.
func BaseTestConfig() (*ini.File, *os.File, error) {
	keyFile, err := generateRSAKeyFile()
	if err != nil {
		return nil, keyFile, err
	}
	// Build ini
	cfg := ini.Empty()
	section, _ := cfg.NewSection("DEFAULT")
	section.NewKey("region", "us-ashburn-1")
	section.NewKey("tenancy", "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	section.NewKey("user", "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	section.NewKey("fingerprint", "3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55")
	section.NewKey("key_file", keyFile.Name())

	return cfg, keyFile, nil
}

// WriteTestConfig writes a ini.File to a temporary file for use in unit tests.
// NOTE: Caller is responsible for removing temporary file.
func WriteTestConfig(cfg *ini.File) (*os.File, error) {
	confFile, err := ioutil.TempFile("", "config_file")
	if err != nil {
		return nil, err
	}

	_, err = cfg.WriteTo(confFile)
	if err != nil {
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

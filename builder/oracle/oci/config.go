package oci

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	ocicommon "github.com/oracle/oci-go-sdk/common"

	"github.com/mitchellh/go-homedir"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	ConfigProvider ocicommon.ConfigurationProvider

	AccessCfgFile        string `mapstructure:"access_cfg_file"`
	AccessCfgFileAccount string `mapstructure:"access_cfg_file_account"`

	// Access config overrides
	UserID       string `mapstructure:"user_ocid"`
	TenancyID    string `mapstructure:"tenancy_ocid"`
	Region       string `mapstructure:"region"`
	Fingerprint  string `mapstructure:"fingerprint"`
	KeyFile      string `mapstructure:"key_file"`
	PassPhrase   string `mapstructure:"pass_phrase"`
	UsePrivateIP bool   `mapstructure:"use_private_ip"`

	AvailabilityDomain string `mapstructure:"availability_domain"`
	CompartmentID      string `mapstructure:"compartment_ocid"`

	// Image
	BaseImageID string `mapstructure:"base_image_ocid"`
	Shape       string `mapstructure:"shape"`
	ImageName   string `mapstructure:"image_name"`

	// Instance
	InstanceName string `mapstructure:"instance_name"`

	// Metadata optionally contains custom metadata key/value pairs provided in the
	// configuration. While this can be used to set metadata["user_data"] the explicit
	// "user_data" and "user_data_file" values will have precedence.
	// An instance's metadata can be obtained from at http://169.254.169.254 on the
	// launched instance.
	Metadata map[string]string `mapstructure:"metadata"`

	// UserData and UserDataFile file are both optional and mutually exclusive.
	UserData     string `mapstructure:"user_data"`
	UserDataFile string `mapstructure:"user_data_file"`

	// Networking
	SubnetID string `mapstructure:"subnet_ocid"`

	// Tagging
	Tags map[string]string `mapstructure:"tags"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, error) {
	c := &Config{}

	// Decode from template
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, fmt.Errorf("Failed to mapstructure Config: %+v", err)
	}

	// Determine where the SDK config is located
	if c.AccessCfgFile == "" {
		c.AccessCfgFile, err = getDefaultOCISettingsPath()
		if err != nil {
			log.Println("Default OCI settings file not found")
		}
	}

	if c.AccessCfgFileAccount == "" {
		c.AccessCfgFileAccount = "DEFAULT"
	}

	var keyContent []byte
	if c.KeyFile != "" {
		path, err := homedir.Expand(c.KeyFile)
		if err != nil {
			return nil, err
		}

		// Read API signing key
		keyContent, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	fileProvider, _ := ocicommon.ConfigurationProviderFromFileWithProfile(c.AccessCfgFile, c.AccessCfgFileAccount, c.PassPhrase)
	if c.Region == "" {
		var region string
		if fileProvider != nil {
			region, _ = fileProvider.Region()
		}
		if region == "" {
			c.Region = "us-phoenix-1"
		}
	}

	providers := []ocicommon.ConfigurationProvider{
		NewRawConfigurationProvider(c.TenancyID, c.UserID, c.Region, c.Fingerprint, string(keyContent), &c.PassPhrase),
	}

	if fileProvider != nil {
		providers = append(providers, fileProvider)
	}

	// Load API access configuration from SDK
	configProvider, err := ocicommon.ComposingConfigurationProvider(providers)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if userOCID, _ := configProvider.UserOCID(); userOCID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'user_ocid' must be specified"))
	}

	tenancyOCID, _ := configProvider.TenancyOCID()
	if tenancyOCID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'tenancy_ocid' must be specified"))
	}

	if fingerprint, _ := configProvider.KeyFingerprint(); fingerprint == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'fingerprint' must be specified"))
	}

	if _, err := configProvider.PrivateRSAKey(); err != nil {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'key_file' must be specified"))
	}

	c.ConfigProvider = configProvider

	if c.AvailabilityDomain == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'availability_domain' must be specified"))
	}

	if c.CompartmentID == "" && tenancyOCID != "" {
		c.CompartmentID = tenancyOCID
	}

	if c.Shape == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'shape' must be specified"))
	}

	if c.SubnetID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'subnet_ocid' must be specified"))
	}

	if c.BaseImageID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'base_image_ocid' must be specified"))
	}

	// Validate tag lengths. TODO (hlowndes) maximum number of tags allowed.
	if c.Tags != nil {
		for k, v := range c.Tags {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if len(k) > 100 {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Tag key length too long. Maximum 100 but found %d. Key: %s", len(k), k))
			}
			if len(k) == 0 {
				errs = packer.MultiErrorAppend(
					errs, errors.New("Tag key empty in config"))
			}
			if len(v) > 100 {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Tag value length too long. Maximum 100 but found %d. Key: %s", len(v), k))
			}
			if len(v) == 0 {
				errs = packer.MultiErrorAppend(
					errs, errors.New("Tag value empty in config"))
			}
		}
	}

	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("unable to parse image name: %s", err))
		} else {
			c.ImageName = name
		}
	}

	// Optional UserData config
	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}
	// read UserDataFile into string.
	if c.UserDataFile != "" {
		fiData, err := ioutil.ReadFile(c.UserDataFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Problem reading user_data_file: %s", err))
		}
		c.UserData = string(fiData)
	}
	// Test if UserData is encoded already, and if not, encode it
	if c.UserData != "" {
		if _, err := base64.StdEncoding.DecodeString(c.UserData); err != nil {
			log.Printf("[DEBUG] base64 encoding user data...")
			c.UserData = base64.StdEncoding.EncodeToString([]byte(c.UserData))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return c, nil
}

// getDefaultOCISettingsPath uses mitchellh/go-homedir to compute the default
// config file location ($HOME/.oci/config).
func getDefaultOCISettingsPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".oci", "config")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	return path, nil
}

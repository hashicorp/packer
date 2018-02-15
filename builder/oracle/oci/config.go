package oci

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	"github.com/mitchellh/go-homedir"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	AccessCfg *client.Config

	AccessCfgFile        string `mapstructure:"access_cfg_file"`
	AccessCfgFileAccount string `mapstructure:"access_cfg_file_account"`

	// Access config overrides
	UserID      string `mapstructure:"user_ocid"`
	TenancyID   string `mapstructure:"tenancy_ocid"`
	Region      string `mapstructure:"region"`
	Fingerprint string `mapstructure:"fingerprint"`
	KeyFile     string `mapstructure:"key_file"`
	PassPhrase  string `mapstructure:"pass_phrase"`

	AvailabilityDomain string `mapstructure:"availability_domain"`
	CompartmentID      string `mapstructure:"compartment_ocid"`

	// Image
	BaseImageID string `mapstructure:"base_image_ocid"`
	Shape       string `mapstructure:"shape"`
	ImageName   string `mapstructure:"image_name"`

	// Networking
	SubnetID string `mapstructure:"subnet_ocid"`

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
	var accessCfgFile string
	if c.AccessCfgFile != "" {
		accessCfgFile = c.AccessCfgFile
	} else {
		accessCfgFile, err = getDefaultOCISettingsPath()
		if err != nil {
			accessCfgFile = "" // Access cfg might be in template
		}
	}

	accessCfg := &client.Config{}

	if accessCfgFile != "" {
		loadedAccessCfgs, err := client.LoadConfigsFromFile(accessCfgFile)
		if err != nil {
			return nil, fmt.Errorf("Invalid config file %s: %s", accessCfgFile, err)
		}
		cfgAccount := "DEFAULT"
		if c.AccessCfgFileAccount != "" {
			cfgAccount = c.AccessCfgFileAccount
		}

		var ok bool
		accessCfg, ok = loadedAccessCfgs[cfgAccount]
		if !ok {
			return nil, fmt.Errorf("No account section '%s' found in config file %s", cfgAccount, accessCfgFile)
		}
	}

	// Override SDK client config with any non-empty template properties

	if c.UserID != "" {
		accessCfg.User = c.UserID
	}

	if c.TenancyID != "" {
		accessCfg.Tenancy = c.TenancyID
	}

	if c.Region != "" {
		accessCfg.Region = c.Region
	}

	// Default if the template nor the API config contains a region.
	if accessCfg.Region == "" {
		accessCfg.Region = "us-phoenix-1"
	}

	if c.Fingerprint != "" {
		accessCfg.Fingerprint = c.Fingerprint
	}

	if c.PassPhrase != "" {
		accessCfg.PassPhrase = c.PassPhrase
	}

	if c.KeyFile != "" {
		accessCfg.KeyFile = c.KeyFile
		accessCfg.Key, err = client.LoadPrivateKey(accessCfg)
		if err != nil {
			return nil, fmt.Errorf("Failed to load private key %s : %s", accessCfg.KeyFile, err)
		}
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Required AccessCfg configuration options

	if accessCfg.User == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'user_ocid' must be specified"))
	}

	if accessCfg.Tenancy == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'tenancy_ocid' must be specified"))
	}

	if accessCfg.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'region' must be specified"))
	}

	if accessCfg.Fingerprint == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'fingerprint' must be specified"))
	}

	if accessCfg.Key == nil {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'key_file' must be specified"))
	}

	c.AccessCfg = accessCfg

	// Required non AccessCfg configuration options

	if c.AvailabilityDomain == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'availability_domain' must be specified"))
	}

	if c.CompartmentID == "" {
		c.CompartmentID = accessCfg.Tenancy
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

	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("unable to parse image name: %s", err))
		} else {
			c.ImageName = name
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

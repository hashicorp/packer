package export

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	ocicommon "github.com/oracle/oci-go-sdk/common"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ConfigProvider ocicommon.ConfigurationProvider

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

	//Object Storage
	BucketName string `mapstructure:"bucket_name"`
	ImageName  string `mapstructure:"image_name"`

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
		path, err := packer.ExpandUser(c.KeyFile)
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

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return c, nil
}

// getDefaultOCISettingsPath uses os/user to compute the default
// config file location ($HOME/.oci/config).
func getDefaultOCISettingsPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	if u.HomeDir == "" {
		return "", fmt.Errorf("Unable to determine the home directory for the current user.")
	}

	path := filepath.Join(u.HomeDir, ".oci", "config")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	return path, nil
}

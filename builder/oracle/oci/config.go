//go:generate mapstructure-to-hcl2 -type Config,CreateVNICDetails,ListImagesRequest

package oci

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	ocicommon "github.com/oracle/oci-go-sdk/common"
	ociauth "github.com/oracle/oci-go-sdk/common/auth"
)

type CreateVNICDetails struct {
	// fields that can be specified under "create_vnic_details"
	AssignPublicIp      *bool                             `mapstructure:"assign_public_ip" required:"false"`
	DefinedTags         map[string]map[string]interface{} `mapstructure:"defined_tags" required:"false"`
	DisplayName         *string                           `mapstructure:"display_name" required:"false"`
	FreeformTags        map[string]string                 `mapstructure:"tags" required:"false"`
	HostnameLabel       *string                           `mapstructure:"hostname_label" required:"false"`
	NsgIds              []string                          `mapstructure:"nsg_ids" required:"false"`
	PrivateIp           *string                           `mapstructure:"private_ip" required:"false"`
	SkipSourceDestCheck *bool                             `mapstructure:"skip_source_dest_check" required:"false"`
	SubnetId            *string                           `mapstructure:"subnet_id" required:"false"`
}

type ListImagesRequest struct {
	// fields that can be specified under "base_image_filter"
	CompartmentId          *string `mapstructure:"compartment_id"`
	DisplayName            *string `mapstructure:"display_name"`
	DisplayNameSearch      *string `mapstructure:"display_name_search"`
	OperatingSystem        *string `mapstructure:"operating_system"`
	OperatingSystemVersion *string `mapstructure:"operating_system_version"`
	Shape                  *string `mapstructure:"shape"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	configProvider ocicommon.ConfigurationProvider

	// Instance Principals (OPTIONAL)
	// If set to true the following can't have non empty values
	// - AccessCfgFile
	// - AccessCfgFileAccount
	// - UserID
	// - TenancyID
	// - Region
	// - Fingerprint
	// - KeyFile
	// - PassPhrase
	InstancePrincipals bool `mapstructure:"use_instance_principals"`

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
	BaseImageID        string            `mapstructure:"base_image_ocid"`
	BaseImageFilter    ListImagesRequest `mapstructure:"base_image_filter"`
	ImageName          string            `mapstructure:"image_name"`
	ImageCompartmentID string            `mapstructure:"image_compartment_ocid"`
	LaunchMode         string            `mapstructure:"image_launch_mode"`

	// Instance
	InstanceName        *string                           `mapstructure:"instance_name"`
	InstanceTags        map[string]string                 `mapstructure:"instance_tags"`
	InstanceDefinedTags map[string]map[string]interface{} `mapstructure:"instance_defined_tags"`
	Shape               string                            `mapstructure:"shape"`
	BootVolumeSizeInGBs int64                             `mapstructure:"disk_size"`

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
	SubnetID          string            `mapstructure:"subnet_ocid"`
	CreateVnicDetails CreateVNICDetails `mapstructure:"create_vnic_details"`

	// Tagging
	Tags        map[string]string                 `mapstructure:"tags"`
	DefinedTags map[string]map[string]interface{} `mapstructure:"defined_tags"`

	ctx interpolate.Context
}

func (c *Config) ConfigProvider() ocicommon.ConfigurationProvider {
	return c.configProvider
}

func (c *Config) Prepare(raws ...interface{}) error {

	// Decode from template
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return fmt.Errorf("Failed to mapstructure Config: %+v", err)
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	var tenancyOCID string

	if c.InstancePrincipals {
		// We could go through all keys in one go and report that the below set
		// of keys cannot coexist with use_instance_principals but decided to
		// split them and report them seperately so that the user sees the specific
		// key involved.
		var message string = " cannot be present when use_instance_principals is set to true."
		if c.AccessCfgFile != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("access_cfg_file"+message))
		}
		if c.AccessCfgFileAccount != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("access_cfg_file_account"+message))
		}
		if c.UserID != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("user_ocid"+message))
		}
		if c.TenancyID != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("tenancy_ocid"+message))
		}
		if c.Region != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("region"+message))
		}
		if c.Fingerprint != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("fingerprint"+message))
		}
		if c.KeyFile != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("key_file"+message))
		}
		if c.PassPhrase != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("pass_phrase"+message))
		}
		// This check is used to facilitate testing. During testing a Mock struct
		// is assigned to c.configProvider otherwise testing fails because Instance
		// Principals cannot be obtained.
		if c.configProvider == nil {
			// Even though the previous configuraion checks might fail we don't want
			// to skip this step. It seems that the logic behind the checks in this
			// file is to check everything even getting the configProvider.
			c.configProvider, err = ociauth.InstancePrincipalConfigurationProvider()
			if err != nil {
				return err
			}
		}
		tenancyOCID, err = c.configProvider.TenancyOCID()
		if err != nil {
			return err
		}
	} else {
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
			path, err := pathing.ExpandUser(c.KeyFile)
			if err != nil {
				return err
			}

			// Read API signing key
			keyContent, err = ioutil.ReadFile(path)
			if err != nil {
				return err
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
			return err
		}

		if userOCID, _ := configProvider.UserOCID(); userOCID == "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("'user_ocid' must be specified"))
		}

		tenancyOCID, _ = configProvider.TenancyOCID()
		if tenancyOCID == "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("'tenancy_ocid' must be specified"))
		}

		if fingerprint, _ := configProvider.KeyFingerprint(); fingerprint == "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("'fingerprint' must be specified"))
		}

		if _, err := configProvider.PrivateRSAKey(); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("'key_file' must be specified"))
		}

		c.configProvider = configProvider
	}

	if c.AvailabilityDomain == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'availability_domain' must be specified"))
	}

	if c.CompartmentID == "" && tenancyOCID != "" {
		c.CompartmentID = tenancyOCID
	}

	if c.ImageCompartmentID == "" {
		c.ImageCompartmentID = c.CompartmentID
	}

	if c.Shape == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'shape' must be specified"))
	}

	if (c.SubnetID == "") && (c.CreateVnicDetails.SubnetId == nil) {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'subnet_ocid' must be specified"))
	}

	if c.CreateVnicDetails.SubnetId == nil {
		c.CreateVnicDetails.SubnetId = &c.SubnetID
	} else if (*c.CreateVnicDetails.SubnetId != c.SubnetID) && (c.SubnetID != "") {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'create_vnic_details[subnet]' must match 'subnet_ocid' if both are specified"))
	}

	if (c.BaseImageID == "") && (c.BaseImageFilter == ListImagesRequest{}) {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'base_image_ocid' or 'base_image_filter' must be specified"))
	}

	if c.BaseImageFilter.CompartmentId == nil {
		c.BaseImageFilter.CompartmentId = &c.CompartmentID
	}

	if c.BaseImageFilter.Shape == nil {
		c.BaseImageFilter.Shape = &c.Shape
	}

	// Validate tag lengths. TODO (hlowndes) maximum number of tags allowed.
	if c.Tags != nil {
		for k, v := range c.Tags {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if len(k) > 100 {
				errs = packersdk.MultiErrorAppend(
					errs, fmt.Errorf("Tag key length too long. Maximum 100 but found %d. Key: %s", len(k), k))
			}
			if len(k) == 0 {
				errs = packersdk.MultiErrorAppend(
					errs, errors.New("Tag key empty in config"))
			}
			if len(v) > 100 {
				errs = packersdk.MultiErrorAppend(
					errs, fmt.Errorf("Tag value length too long. Maximum 100 but found %d. Key: %s", len(v), k))
			}
			if len(v) == 0 {
				errs = packersdk.MultiErrorAppend(
					errs, errors.New("Tag value empty in config"))
			}
		}
	}

	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("unable to parse image name: %s", err))
		} else {
			c.ImageName = name
		}
	}

	// Optional UserData config
	if c.UserData != "" && c.UserDataFile != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}
	// read UserDataFile into string.
	if c.UserDataFile != "" {
		fiData, err := ioutil.ReadFile(c.UserDataFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Problem reading user_data_file: %s", err))
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

	// Set default boot volume size to 50 if not set
	// Check if size set is allowed by OCI
	if c.BootVolumeSizeInGBs == 0 {
		c.BootVolumeSizeInGBs = 50
	} else if c.BootVolumeSizeInGBs < 50 || c.BootVolumeSizeInGBs > 16384 {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("'disk_size' must be between 50 and 16384 GBs"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
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

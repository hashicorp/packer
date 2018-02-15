package ncloud

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is structure to use packer builder plugin for Naver Cloud Platform
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccessKey                         string `mapstructure:"access_key"`
	SecretKey                         string `mapstructure:"secret_key"`
	ServerImageProductCode            string `mapstructure:"server_image_product_code"`
	ServerProductCode                 string `mapstructure:"server_product_code"`
	MemberServerImageNo               string `mapstructure:"member_server_image_no"`
	ServerImageName                   string `mapstructure:"server_image_name"`
	ServerImageDescription            string `mapstructure:"server_image_description"`
	UserData                          string `mapstructure:"user_data"`
	UserDataFile                      string `mapstructure:"user_data_file"`
	BlockStorageSize                  int    `mapstructure:"block_storage_size"`
	Region                            string `mapstructure:"region"`
	AccessControlGroupConfigurationNo string `mapstructure:"access_control_group_configuration_no"`

	Comm communicator.Config `mapstructure:",squash"`
	ctx  *interpolate.Context
}

// NewConfig checks parameters
func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	warnings := []string{}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return nil, warnings, err
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(nil); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if c.AccessKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("access_key is required"))
	}

	if c.SecretKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("secret_key is required"))
	}

	if c.MemberServerImageNo == "" && c.ServerImageProductCode == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("server_image_product_code or member_server_image_no is required"))
	}

	if c.MemberServerImageNo != "" && c.ServerImageProductCode != "" {
		errs = packer.MultiErrorAppend(errs, errors.New("Only one of server_image_product_code and member_server_image_no can be set"))
	}

	if c.ServerImageProductCode != "" && len(c.ServerImageProductCode) > 20 {
		errs = packer.MultiErrorAppend(errs, errors.New("If server_image_product_code field is set, length of server_image_product_code should be max 20"))
	}

	if c.ServerProductCode != "" && len(c.ServerProductCode) > 20 {
		errs = packer.MultiErrorAppend(errs, errors.New("If server_product_code field is set, length of server_product_code should be max 20"))
	}

	if c.ServerImageName != "" && (len(c.ServerImageName) < 3 || len(c.ServerImageName) > 30) {
		errs = packer.MultiErrorAppend(errs, errors.New("If server_image_name field is set, length of server_image_name should be min 3 and max 20"))
	}

	if c.ServerImageDescription != "" && len(c.ServerImageDescription) > 1000 {
		errs = packer.MultiErrorAppend(errs, errors.New("If server_image_description field is set, length of server_image_description should be max 1000"))
	}

	if c.BlockStorageSize != 0 {
		if c.BlockStorageSize < 10 || c.BlockStorageSize > 2000 {
			errs = packer.MultiErrorAppend(errs, errors.New("The size of BlockStorageSize is at least 10 GB and up to 2000GB"))
		} else if int(c.BlockStorageSize/10)*10 != c.BlockStorageSize {
			return nil, nil, errors.New("BlockStorageSize must be a multiple of 10 GB")
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(errs, errors.New("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.UserData != "" && len(c.UserData) > 21847 {
		errs = packer.MultiErrorAppend(errs, errors.New("If user_data field is set, length of UserData should be max 21847"))
	}

	if c.Comm.Type == "wrinrm" && c.AccessControlGroupConfigurationNo == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("If Communicator is winrm, access_control_group_configuration_no is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

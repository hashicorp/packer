//go:generate struct-markdown

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
	// Product code of an image to create.
    // (member_server_image_no is required if not specified)
	ServerImageProductCode            string `mapstructure:"server_image_product_code" required:"true"`
	// Product (spec) code to create.
	ServerProductCode                 string `mapstructure:"server_product_code" required:"true"`
	// Previous image code. If there is an
    // image previously created, it can be used to create a new image.
    // (server_image_product_code is required if not specified)
	MemberServerImageNo               string `mapstructure:"member_server_image_no" required:"false"`
	// Name of an image to create.
	ServerImageName                   string `mapstructure:"server_image_name" required:"false"`
	// Description of an image to create.
	ServerImageDescription            string `mapstructure:"server_image_description" required:"false"`
	// User data to apply when launching the instance. Note
    // that you need to be careful about escaping characters due to the templates
    // being JSON. It is often more convenient to use user_data_file, instead.
    // Packer will not automatically wait for a user script to finish before
    // shutting down the instance this must be handled in a provisioner.
	UserData                          string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
    // data when launching the instance.
	UserDataFile                      string `mapstructure:"user_data_file" required:"false"`
	// You can add block storage ranging from 10
    // GB to 2000 GB, in increments of 10 GB.
	BlockStorageSize                  int    `mapstructure:"block_storage_size" required:"false"`
	// Name of the region where you want to create an image.
    // (default: Korea)
	Region                            string `mapstructure:"region" required:"false"`
	// This is used to allow
    // winrm access when you create a Windows server. An ACG that specifies an
    // access source (0.0.0.0/0) and allowed port (5985) must be created in
    // advance.
	AccessControlGroupConfigurationNo string `mapstructure:"access_control_group_configuration_no" required:"false"`

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

	if c.Comm.Type == "winrm" && c.AccessControlGroupConfigurationNo == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("If Communicator is winrm, access_control_group_configuration_no is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

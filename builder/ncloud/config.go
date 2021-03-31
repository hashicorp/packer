//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package ncloud

import (
	"errors"
	"fmt"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"os"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Config is structure to use packer builder plugin for Naver Cloud Platform
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	// Product code of an image to create.
	// (member_server_image_no is required if not specified)
	ServerImageProductCode string `mapstructure:"server_image_product_code" required:"true"`
	// Product (spec) code to create.
	ServerProductCode string `mapstructure:"server_product_code" required:"true"`
	// Previous image code. If there is an
	// image previously created, it can be used to create a new image.
	// (server_image_product_code is required if not specified)
	MemberServerImageNo string `mapstructure:"member_server_image_no" required:"false"`
	// Name of an image to create.
	ServerImageName string `mapstructure:"server_image_name" required:"false"`
	// Description of an image to create.
	ServerImageDescription string `mapstructure:"server_image_description" required:"false"`
	// User data to apply when launching the instance. Note
	// that you need to be careful about escaping characters due to the templates
	// being JSON. It is often more convenient to use user_data_file, instead.
	// Packer will not automatically wait for a user script to finish before
	// shutting down the instance this must be handled in a provisioner.
	UserData string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
	// data when launching the instance.
	UserDataFile string `mapstructure:"user_data_file" required:"false"`
	// You can add block storage ranging from 10
	// GB to 2000 GB, in increments of 10 GB.
	BlockStorageSize int `mapstructure:"block_storage_size" required:"false"`
	// Name of the region where you want to create an image.
	// (default: Korea)
	Region     string `mapstructure:"region" required:"false"`
	RegionCode string `mapstructure:"region_code" required:"false"`
	// Deprecated
	AccessControlGroupConfigurationNo string `mapstructure:"access_control_group_configuration_no" required:"false"`
	// This is used to allow
	// winrm access when you create a Windows server. An ACG that specifies an
	// access source (0.0.0.0/0) and allowed port (5985) must be created in
	// advance if you use CLASSIC env. If this field is left blank,
	// Packer will create temporary ACG for automatically in VPC environment.
	AccessControlGroupNo string `mapstructure:"access_control_group_no" required:"false"`
	SupportVPC bool `mapstructure:"support_vpc" required:"false"`
	// The ID of the Subnet where you want to place the Server Instance. If this field is left blank, Packer will try to get the Public Subnet ID from the `vpc_no`.
	SubnetNo string `mapstructure:"subnet_no" required:"false"`
	// The ID of the VPC where you want to place the Server Instance. If this field is left blank, Packer will try to get the VPC ID from the `subnet_no`.
	// (You are required to least one between two parameters if u want using VPC environment: `vpc_no` or `subnet_no`)
	VpcNo string `mapstructure:"vpc_no" required:"false"`

	Comm communicator.Config `mapstructure:",squash"`
	ctx  *interpolate.Context
}

// NewConfig checks parameters
func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	var warnings []string

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return warnings, err
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(nil); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	if c.AccessKey == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("`access_key` is required"))
	}

	if c.SecretKey == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("`secret_key` is required"))
	}

	if c.MemberServerImageNo == "" && c.ServerImageProductCode == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("`server_image_product_code` or `member_server_image_no` is required"))
	}

	if c.MemberServerImageNo != "" && c.ServerImageProductCode != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("only one of `server_image_product_code` and `member_server_image_no` can be set"))
	}

	if c.ServerImageProductCode != "" && len(c.ServerImageProductCode) > 50 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if `server_image_product_code` field is set, length of `server_image_product_code` should be max 20"))
	}

	if c.ServerProductCode != "" && len(c.ServerProductCode) > 50 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if `server_product_code` field is set, length of `server_product_code` should be max 20"))
	}

	if c.ServerImageName != "" && (len(c.ServerImageName) < 3 || len(c.ServerImageName) > 30) {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if `server_image_name` field is set, length of `server_image_name` should be min 3 and max 20"))
	}

	if c.ServerImageDescription != "" && len(c.ServerImageDescription) > 1000 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if `server_image_description` field is set, length of `server_image_description` should be max 1000"))
	}

	if c.BlockStorageSize != 0 {
		if c.BlockStorageSize < 10 || c.BlockStorageSize > 2000 {
			errs = packersdk.MultiErrorAppend(errs, errors.New("the size of `block_storage_size` is at least 10 GB and up to 2000GB"))
		} else if int(c.BlockStorageSize/10)*10 != c.BlockStorageSize {
			return nil, errors.New("BlockStorageSize must be a multiple of 10 GB")
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("only one of user_data or `user_data`_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`user_data_file` not found: %s", c.UserDataFile))
		}
	}

	if c.UserData != "" && len(c.UserData) > 21847 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if `user_data` field is set, length of UserData should be max 21847"))
	}

	if c.AccessControlGroupConfigurationNo != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("`access_control_group_configuration_no` is deprecated, please use `access_control_group_no` instead"))
	}

	if c.VpcNo != "" || c.SubnetNo != "" {
		c.SupportVPC = true
	}

	if c.Comm.Type == "winrm" && c.AccessControlGroupNo == "" && !c.SupportVPC {
		errs = packersdk.MultiErrorAppend(errs, errors.New("if Communicator is winrm, `access_control_group_no` (allow 5986 port) is required in `CLASSIC` environment"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

type NcloudAPIClient struct {
	server  *server.APIClient
	vserver *vserver.APIClient
	vpc     *vpc.APIClient
}

func (c *Config) Client() (*NcloudAPIClient, error) {
	apiKey := &ncloud.APIKey{
		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	}
	return &NcloudAPIClient{
		server:  server.NewAPIClient(server.NewConfiguration(apiKey)),
		vserver: vserver.NewAPIClient(vserver.NewConfiguration(apiKey)),
		vpc:     vpc.NewAPIClient(vpc.NewConfiguration(apiKey)),
	}, nil
}

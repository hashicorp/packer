//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package yandex

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Communicator        communicator.Config `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`

	CommonConfig `mapstructure:",squash"`
	ImageConfig  `mapstructure:",squash"`

	SourceImageConfig `mapstructure:",squash"`
	// Service account identifier to assign to instance.
	ServiceAccountID string `mapstructure:"service_account_id" required:"false"`

	// The ID of the folder to save built image in.
	// This defaults to value of 'folder_id'.
	TargetImageFolderID string `mapstructure:"target_image_folder_id" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, c.AccessConfig.Prepare(&c.ctx)...)

	errs = c.CommonConfig.Prepare(errs)
	errs = c.ImageConfig.Prepare(errs)
	errs = c.SourceImageConfig.Prepare(errs)

	if c.ImageMinDiskSizeGb == 0 {
		c.ImageMinDiskSizeGb = c.DiskSizeGb
	}

	if c.ImageMinDiskSizeGb < c.DiskSizeGb {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("Invalid image_min_disk_size value (%d): Must be equal or greate than disk_size_gb (%d)",
				c.ImageMinDiskSizeGb, c.DiskSizeGb))
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName + "-disk"
	}

	if es := c.Communicator.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	if c.TargetImageFolderID == "" {
		c.TargetImageFolderID = c.FolderID
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

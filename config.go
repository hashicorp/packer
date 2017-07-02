package main

import (
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig 	`mapstructure:",squash"`
	ConnectConfig 			`mapstructure:",squash"`
	CloneConfig 			`mapstructure:",squash"`
	HardwareConfig 			`mapstructure:",squash"`
	communicator.Config 	`mapstructure:",squash"`
	ShutdownConfig         	`mapstructure:",squash"`
	CreateSnapshot    bool 	`mapstructure:"create_snapshot"`
	ConvertToTemplate bool 	`mapstructure:"convert_to_template"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	{
		err := config.Decode(c, &config.DecodeOpts{
			Interpolate:        true,
			InterpolateContext: &c.ctx,
		}, raws...)
		if err != nil {
			return nil, nil, err
		}
	}

	errs := new(packer.MultiError)
	errs = packer.MultiErrorAppend(errs, c.Config.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ConnectConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.CloneConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.HardwareConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare()...)

	if len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

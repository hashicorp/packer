package common

import (
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template/interpolate"
)

func DecodeConfig(cfg interface{}, ctx *interpolate.Context, raws ...interface{}) error {
	err := config.Decode(cfg, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: ctx,
	}, raws...)
	return err
}

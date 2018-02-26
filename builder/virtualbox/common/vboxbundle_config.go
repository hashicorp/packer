package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxBundleConfig struct {
  BundleISO      bool `mapstructure:"bundle_iso"`
}

func (c *VBoxBundleConfig) Prepare(ctx *interpolate.Context) []error {
	return nil
}

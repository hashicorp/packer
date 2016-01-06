package common

import (
	"os"

	"github.com/mitchellh/packer/template/interpolate"
)

type DriverConfig struct {
	FusionAppPath string `mapstructure:"fusion_app_path"`
}

func (c *DriverConfig) Prepare(ctx *interpolate.Context) []error {
	if c.FusionAppPath == "" {
		c.FusionAppPath = os.Getenv("FUSION_APP_PATH")
	}
	if c.FusionAppPath == "" {
		c.FusionAppPath = "/Applications/VMware Fusion.app"
	}

	return nil
}

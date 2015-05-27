package openstack

import (
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// ImageConfig is for common configuration related to creating Images.
type ImageConfig struct {
	ImageName string `mapstructure:"image_name"`
}

func (c *ImageConfig) Prepare(ctx *interpolate.Context) []error {
	errs := make([]error, 0)
	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("An image_name must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

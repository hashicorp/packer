package openstack

import (
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// ImageConfig is for common configuration related to creating Images.
type ImageConfig struct {
	ImageName     string            `mapstructure:"image_name"`
	ImageMetadata map[string]string `mapstructure:"metadata"`
}

func (c *ImageConfig) Prepare(ctx *interpolate.Context) []error {
	errs := make([]error, 0)
	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("An image_name must be specified"))
	}

	// By default, OpenStack seems to create the image with an image_type of
	// "snapshot", since it came from snapshotting a VM. A "snapshot" looks
	// slightly different in the OpenStack UI and OpenStack won't show "snapshot"
	// images as a choice in the list of images to boot from for a new instance.
	// See https://github.com/mitchellh/packer/issues/3038
	if c.ImageMetadata == nil {
		c.ImageMetadata = map[string]string{"image_type": "image"}
	} else if c.ImageMetadata["image_type"] == "" {
		c.ImageMetadata["image_type"] = "image"
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

package openstack

import (
	"fmt"

	imageservice "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/mitchellh/packer/template/interpolate"
)

// ImageConfig is for common configuration related to creating Images.
type ImageConfig struct {
	ImageName string `mapstructure:"image_name"`

	ImageMetadata   map[string]string            `mapstructure:"metadata"`
	ImageVisibility imageservice.ImageVisibility `mapstructure:"image_visibility"`
	ImageMembers    []string                     `mapstructure:"image_members"`
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

	// ImageVisibility values
	// https://wiki.openstack.org/wiki/Glance-v2-community-image-visibility-design
	if c.ImageVisibility != "" {
		validVals := []string{"public", "private", "shared", "community"}
		valid := false
		for _, val := range validVals {
			if string(c.ImageVisibility) == val {
				valid = true
				break
			}
		}
		if !valid {
			errs = append(errs, fmt.Errorf("Unknown visibility value %s", c.ImageVisibility))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

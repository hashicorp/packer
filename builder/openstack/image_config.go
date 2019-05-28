package openstack

import (
	"fmt"
	"strings"

	imageservice "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/template/interpolate"
)

// ImageConfig is for common configuration related to creating Images.
type ImageConfig struct {
	// The name of the resulting image.
	ImageName       string                       `mapstructure:"image_name" required:"true"`
	// Glance metadata that will be
    // applied to the image.
	ImageMetadata   map[string]string            `mapstructure:"metadata" required:"false"`
	// One of "public", "private", "shared", or
    // "community".
	ImageVisibility imageservice.ImageVisibility `mapstructure:"image_visibility" required:"false"`
	// List of members to add to the image
    // after creation. An image member is usually a project (also called the
    // "tenant") with whom the image is shared.
	ImageMembers    []string                     `mapstructure:"image_members" required:"false"`
	// Disk format of the resulting image. This
    // option works if use_blockstorage_volume is true.
	ImageDiskFormat string                       `mapstructure:"image_disk_format" required:"false"`
	// List of tags to add to the image after
    // creation.
	ImageTags       []string                     `mapstructure:"image_tags" required:"false"`
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
	// See https://github.com/hashicorp/packer/issues/3038
	if c.ImageMetadata == nil {
		c.ImageMetadata = map[string]string{"image_type": "image"}
	} else if c.ImageMetadata["image_type"] == "" {
		c.ImageMetadata["image_type"] = "image"
	}

	// ImageVisibility values
	// https://wiki.openstack.org/wiki/Glance-v2-community-image-visibility-design
	if c.ImageVisibility != "" {
		validVals := []imageservice.ImageVisibility{"public", "private", "shared", "community"}
		valid := false
		for _, val := range validVals {
			if strings.EqualFold(string(c.ImageVisibility), string(val)) {
				valid = true
				c.ImageVisibility = val
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

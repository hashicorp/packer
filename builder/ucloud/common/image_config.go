//go:generate mapstructure-to-hcl2 -type ImageDestination

package common

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

type ImageDestination struct {
	ProjectId   string `mapstructure:"project_id"`
	Region      string `mapstructure:"region"`
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

type ImageConfig struct {
	ImageName             string             `mapstructure:"image_name"`
	ImageDescription      string             `mapstructure:"image_description"`
	ImageDestinations     []ImageDestination `mapstructure:"image_copy_to_mappings"`
	WaitImageReadyTimeout int                `mapstructure:"wait_image_ready_timeout"`
}

var ImageNamePattern = regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_\[\]:,.]{1,63}$`)

func (c *ImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	imageName := c.ImageName
	if imageName == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "image_name"))
	} else if !ImageNamePattern.MatchString(imageName) {
		errs = append(errs, fmt.Errorf("expected %q to be 1-63 characters and only support chinese, english, numbers, '-_,.:[]', got %q", "image_name", imageName))
	}

	if len(c.ImageDestinations) > 0 {
		for _, imageDestination := range c.ImageDestinations {
			if imageDestination.Name == "" {
				imageDestination.Name = imageName
			}

			errs = append(errs, imageDestination.validate()...)
		}
	}

	if c.WaitImageReadyTimeout <= 0 {
		c.WaitImageReadyTimeout = DefaultCreateImageTimeOut
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (imageDestination *ImageDestination) validate() []error {
	var errs []error

	if imageDestination.Region == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "image_copy_region"))
	}

	if imageDestination.ProjectId == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "image_copy_project"))
	}

	if imageDestination.Name != "" && !ImageNamePattern.MatchString(imageDestination.Name) {
		errs = append(errs, fmt.Errorf("expected %q to be 1-63 characters and only support chinese, english, numbers, '-_,.:[]', got %q", "image_copy_name", imageDestination.Name))
	}

	return errs
}

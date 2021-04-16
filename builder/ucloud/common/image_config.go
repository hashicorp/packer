//go:generate packer-sdc mapstructure-to-hcl2 -type ImageDestination
//go:generate packer-sdc struct-markdown

package common

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type ImageDestination struct {
	// The destination project id, where copying image in.
	ProjectId string `mapstructure:"project_id" required:"false"`
	// The destination region, where copying image in.
	Region string `mapstructure:"region" required:"false"`
	// The copied image name. If not defined, builder will use `image_name` as default name.
	Name string `mapstructure:"name" required:"false"`
	// The copied image description.
	Description string `mapstructure:"description" required:"false"`
}

type ImageConfig struct {
	// The name of the user-defined image, which contains 1-63 characters and only
	// support Chinese, English, numbers, '-\_,.:[]'.
	ImageName string `mapstructure:"image_name" required:"true"`
	// The description of the image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	// The array of mappings regarding the copied images to the destination regions and projects.
	//
	//  - `project_id` (string) - The destination project id, where copying image in.
	//
	//  - `region` (string) - The destination region, where copying image in.
	//
	//  - `name` (string) - The copied image name. If not defined, builder will use `image_name` as default name.
	//
	//  - `description` (string) - The copied image description.
	//
	// ```json
	// {
	//   "image_copy_to_mappings": [
	//     {
	//       "project_id": "{{user `ucloud_project_id`}}",
	//       "region": "cn-sh2",
	//       "description": "test",
	//       "name": "packer-test-basic-sh"
	//     }
	//   ]
	// }
	// ```
	ImageDestinations []ImageDestination `mapstructure:"image_copy_to_mappings" required:"false"`
	// Timeout of creating image or copying image. The default timeout is 3600 seconds if this option
	// is not set or is set to 0.
	WaitImageReadyTimeout int `mapstructure:"wait_image_ready_timeout" required:"false"`
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
		c.WaitImageReadyTimeout = DefaultCreateImageTimeout
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

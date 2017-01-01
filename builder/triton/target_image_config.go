package triton

import (
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// TargetImageConfig represents the configuration for the image to be created
// from the source machine.
type TargetImageConfig struct {
	ImageName        string            `mapstructure:"image_name"`
	ImageVersion     string            `mapstructure:"image_version"`
	ImageDescription string            `mapstructure:"image_description"`
	ImageHomepage    string            `mapstructure:"image_homepage"`
	ImageEULA        string            `mapstructure:"image_eula_url"`
	ImageACL         []string          `mapstructure:"image_acls"`
	ImageTags        map[string]string `mapstructure:"image_tags"`
}

// Prepare performs basic validation on a TargetImageConfig struct.
func (c *TargetImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("An image_name must be specified"))
	}

	if c.ImageVersion == "" {
		errs = append(errs, fmt.Errorf("An image_version must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

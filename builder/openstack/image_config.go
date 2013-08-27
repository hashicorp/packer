package openstack

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

// ImageConfig is for common configuration related to creating Images.
type ImageConfig struct {
	ImageName string `mapstructure:"image_name"`
}

func (c *ImageConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	templates := map[string]*string{
		"image_name": &c.ImageName,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("An image_name must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

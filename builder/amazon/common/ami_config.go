package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/packer/packer"
)

// AMIConfig is for common configuration related to creating AMIs.
type AMIConfig struct {
	AMIName         string   `mapstructure:"ami_name"`
	AMIDescription  string   `mapstructure:"ami_description"`
	AMIUsers        []string `mapstructure:"ami_users"`
	AMIGroups       []string `mapstructure:"ami_groups"`
	AMIProductCodes []string `mapstructure:"ami_product_codes"`
	AMIRegions      []string `mapstructure:"ami_regions"`
}

func (c *AMIConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	templates := map[string]*string{
		"ami_name":        &c.AMIName,
		"ami_description": &c.AMIDescription,
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

	sliceTemplates := map[string][]string{
		"ami_users":         c.AMIUsers,
		"ami_groups":        c.AMIGroups,
		"ami_product_codes": c.AMIProductCodes,
		"ami_regions":       c.AMIRegions,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = t.Process(elem, nil)
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	if c.AMIName == "" {
		errs = append(errs, fmt.Errorf("ami_name must be specified"))
	}

	if len(c.AMIRegions) > 0 {
		for _, region := range c.AMIRegions {
			if _, ok := aws.Regions[region]; !ok {
				errs = append(errs, fmt.Errorf("Unknown region: %s", region))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/packer/packer"
)

// AMIConfig is for common configuration related to creating AMIs.
type AMIConfig struct {
	AMIName         string            `mapstructure:"ami_name"`
	AMIDescription  string            `mapstructure:"ami_description"`
	AMIVirtType     string            `mapstructure:"ami_virtualization_type"`
	AMIUsers        []string          `mapstructure:"ami_users"`
	AMIGroups       []string          `mapstructure:"ami_groups"`
	AMIProductCodes []string          `mapstructure:"ami_product_codes"`
	AMIRegions      []string          `mapstructure:"ami_regions"`
	AMITags         map[string]string `mapstructure:"tags"`
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
		"ami_name":                &c.AMIName,
		"ami_description":         &c.AMIDescription,
		"ami_virtualization_type": &c.AMIVirtType,
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
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(c.AMIRegions))

		for _, region := range c.AMIRegions {
			// If we already saw the region, then don't look again
			if _, ok := regionSet[region]; ok {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = struct{}{}

			// Verify the region is real
			if _, ok := aws.Regions[region]; !ok {
				errs = append(errs, fmt.Errorf("Unknown region: %s", region))
				continue
			}

			regions = append(regions, region)
		}

		c.AMIRegions = regions
	}

	newTags := make(map[string]string)
	for k, v := range c.AMITags {
		k, err := t.Process(k, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing tag key %s: %s", k, err))
			continue
		}

		v, err := t.Process(v, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing tag value '%s': %s", v, err))
			continue
		}

		newTags[k] = v
	}

	c.AMITags = newTags

	if len(errs) > 0 {
		return errs
	}

	return nil
}

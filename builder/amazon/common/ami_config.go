package common

import (
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// AMIConfig is for common configuration related to creating AMIs.
type AMIConfig struct {
	AMIName                 string            `mapstructure:"ami_name"`
	AMIDescription          string            `mapstructure:"ami_description"`
	AMIVirtType             string            `mapstructure:"ami_virtualization_type"`
	AMIUsers                []string          `mapstructure:"ami_users"`
	AMIGroups               []string          `mapstructure:"ami_groups"`
	AMIProductCodes         []string          `mapstructure:"ami_product_codes"`
	AMIRegions              []string          `mapstructure:"ami_regions"`
	AMISkipRegionValidation bool              `mapstructure:"skip_region_validation"`
	AMITags                 map[string]string `mapstructure:"tags"`
	AMIEnhancedNetworking   bool              `mapstructure:"enhanced_networking"`
	AMIForceDeregister      bool              `mapstructure:"force_deregister"`
}

func (c *AMIConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
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

			if !c.AMISkipRegionValidation {
				// Verify the region is real
				if valid := ValidateRegion(region); valid == false {
					errs = append(errs, fmt.Errorf("Unknown region: %s", region))
					continue
				}
			}

			regions = append(regions, region)
		}

		c.AMIRegions = regions
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

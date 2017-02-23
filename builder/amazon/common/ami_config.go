package common

import (
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// AMIConfig is for common configuration related to creating AMIs.
type AMIConfig struct {
	AMIName                   string            `mapstructure:"ami_name"`
	AMIDescription            string            `mapstructure:"ami_description"`
	AMIVirtType               string            `mapstructure:"ami_virtualization_type"`
	AMIUsers                  []string          `mapstructure:"ami_users"`
	AMIGroups                 []string          `mapstructure:"ami_groups"`
	AMIProductCodes           []string          `mapstructure:"ami_product_codes"`
	AMIRegions                []string          `mapstructure:"ami_regions"`
	AMISkipRegionValidation   bool              `mapstructure:"skip_region_validation"`
	AMITags                   map[string]string `mapstructure:"tags"`
	AMIEnhancedNetworkingType string            `mapstructure:"enhanced_networking_type"`
	AMIForceDeregister        bool              `mapstructure:"force_deregister"`
	AMIForceDeleteSnapshot    bool              `mapstructure:"force_delete_snapshot"`
	AMIEncryptBootVolume      bool              `mapstructure:"encrypt_boot"`
	AMIKmsKeyId               string            `mapstructure:"kms_key_id"`
	SnapshotTags              map[string]string `mapstructure:"snapshot_tags"`
	SnapshotUsers             []string          `mapstructure:"snapshot_users"`
	SnapshotGroups            []string          `mapstructure:"snapshot_groups"`
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

	if len(c.AMIUsers) > 0 && c.AMIEncryptBootVolume {
		errs = append(errs, fmt.Errorf("Cannot share AMI with encrypted boot volume"))
	}

	if len(c.SnapshotUsers) > 0 && len(c.AMIKmsKeyId) == 0 && c.AMIEncryptBootVolume {
		errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

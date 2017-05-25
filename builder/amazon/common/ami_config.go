package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
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
	AMIForceDeleteSnapshot  bool              `mapstructure:"force_delete_snapshot"`
	AMIEncryptBootVolume    bool              `mapstructure:"encrypt_boot"`
	AMIKmsKeyId             string            `mapstructure:"kms_key_id"`
	AMIRegionKmsKeyIds      map[string]string `mapstructure:"region_kms_key_ids"`
	SnapshotTags            map[string]string `mapstructure:"snapshot_tags"`
	SnapshotUsers           []string          `mapstructure:"snapshot_users"`
	SnapshotGroups          []string          `mapstructure:"snapshot_groups"`
}

func stringInSlice(searchstr string, searchslice []string) bool {
	for _, item := range searchslice {
		if item == searchstr {
			return true
		}
	}
	return false
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
				if valid := ValidateRegion(region); !valid {
					errs = append(errs, fmt.Errorf("Unknown region: %s", region))
					continue
				}
			}

			// Make sure that if we have region_kms_key_ids defined the regions in ami_regions are also in region_kms_key_ids
			if len(c.AMIRegionKmsKeyIds) > 0 {
				regions_in_key_map := make([]string, 0, len(c.AMIRegionKmsKeyIds))
				for reg := range c.AMIRegionKmsKeyIds {
					regions_in_key_map = append(regions_in_key_map, reg)
				}
				if regions_match := stringInSlice(region, regions_in_key_map); !regions_match {
					errs = append(errs, fmt.Errorf("Region %s is in ami_regions but not in region_kms_key_ids", region))
				}
			}

			regions = append(regions, region)
		}

		c.AMIRegions = regions
	}
	// Make sure that if we have region_kms_key_ids defined the regions in region_kms_key_ids are also in ami_regions
	if len(c.AMIRegionKmsKeyIds) > 0 {
		for KMS_key_region := range c.AMIRegionKmsKeyIds {
			if regions_match := stringInSlice(KMS_key_region, c.AMIRegions); !regions_match {
				errs = append(errs, fmt.Errorf("Region %s is in region_kms_key_ids but not in ami_regions", KMS_key_region))
			}
		}
	}

	if len(c.AMIUsers) > 0 && c.AMIEncryptBootVolume {
		errs = append(errs, fmt.Errorf("Cannot share AMI with encrypted boot volume"))
	}

	if len(c.SnapshotUsers) > 0 {
		if len(c.AMIKmsKeyId) == 0 && c.AMIEncryptBootVolume {
			errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
		}
		if len(c.AMIRegionKmsKeyIds) > 0 {
			for _, KMS_key_region := range c.AMIRegionKmsKeyIds {
				if len(KMS_key_region) == 0 {
					errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
				}
			}
		}
	}

	if len(c.AMIName) < 3 || len(c.AMIName) > 128 {
		errs = append(errs, fmt.Errorf("AMIName must be between 3 and 128 characters long"))
	}

	if c.AMIName != templateCleanAMIName(c.AMIName) {
		errs = append(errs, fmt.Errorf("AMIName should only contain "+
			"alphanumeric characters, parentheses (()), square brackets ([]), spaces "+
			"( ), periods (.), slashes (/), dashes (-), single quotes ('), at-signs "+
			"(@), or underscores(_). You can use the `clean_ami_name` template "+
			"filter to automatically clean your ami name."))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

package common

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

//Temporal
type TagMap map[string]string

// OMIConfig is for common configuration related to creating OMIs.
type OMIConfig struct {
	OMIName                 string            `mapstructure:"ami_name"`
	OMIDescription          string            `mapstructure:"ami_description"`
	OMIVirtType             string            `mapstructure:"ami_virtualization_type"`
	OMIUsers                []string          `mapstructure:"ami_users"`
	OMIGroups               []string          `mapstructure:"ami_groups"`
	OMIProductCodes         []string          `mapstructure:"ami_product_codes"`
	OMIRegions              []string          `mapstructure:"ami_regions"`
	OMISkipRegionValidation bool              `mapstructure:"skip_region_validation"`
	OMITags                 TagMap            `mapstructure:"tags"`
	OMIENASupport           *bool             `mapstructure:"ena_support"`
	OMISriovNetSupport      bool              `mapstructure:"sriov_support"`
	OMIForceDeregister      bool              `mapstructure:"force_deregister"`
	OMIForceDeleteSnapshot  bool              `mapstructure:"force_delete_snapshot"`
	OMIEncryptBootVolume    bool              `mapstructure:"encrypt_boot"`
	OMIKmsKeyId             string            `mapstructure:"kms_key_id"`
	OMIRegionKMSKeyIDs      map[string]string `mapstructure:"region_kms_key_ids"`
	SnapshotTags            TagMap            `mapstructure:"snapshot_tags"`
	SnapshotUsers           []string          `mapstructure:"snapshot_users"`
	SnapshotGroups          []string          `mapstructure:"snapshot_groups"`
}

func stringInSlice(s []string, searchstr string) bool {
	for _, item := range s {
		if item == searchstr {
			return true
		}
	}
	return false
}

func (c *OMIConfig) Prepare(accessConfig *AccessConfig, ctx *interpolate.Context) []error {
	var errs []error

	if c.OMIName == "" {
		errs = append(errs, fmt.Errorf("ami_name must be specified"))
	}

	// Make sure that if we have region_kms_key_ids defined,
	//  the regions in region_kms_key_ids are also in ami_regions
	if len(c.OMIRegionKMSKeyIDs) > 0 {
		for kmsKeyRegion := range c.OMIRegionKMSKeyIDs {
			if !stringInSlice(c.OMIRegions, kmsKeyRegion) {
				errs = append(errs, fmt.Errorf("Region %s is in region_kms_key_ids but not in ami_regions", kmsKeyRegion))
			}
		}
	}

	errs = append(errs, c.prepareRegions(accessConfig)...)

	if len(c.OMIUsers) > 0 && c.OMIEncryptBootVolume {
		errs = append(errs, fmt.Errorf("Cannot share OMI with encrypted boot volume"))
	}

	var kmsKeys []string
	if len(c.OMIKmsKeyId) > 0 {
		kmsKeys = append(kmsKeys, c.OMIKmsKeyId)
	}
	if len(c.OMIRegionKMSKeyIDs) > 0 {
		for _, kmsKey := range c.OMIRegionKMSKeyIDs {
			if len(kmsKey) == 0 {
				kmsKeys = append(kmsKeys, c.OMIKmsKeyId)
			}
		}
	}
	for _, kmsKey := range kmsKeys {
		if !validateKmsKey(kmsKey) {
			errs = append(errs, fmt.Errorf("%s is not a valid KMS Key Id.", kmsKey))
		}
	}

	if len(c.SnapshotUsers) > 0 {
		if len(c.OMIKmsKeyId) == 0 && c.OMIEncryptBootVolume {
			errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
		}
		if len(c.OMIRegionKMSKeyIDs) > 0 {
			for _, kmsKey := range c.OMIRegionKMSKeyIDs {
				if len(kmsKey) == 0 {
					errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
				}
			}
		}
	}

	if len(c.OMIName) < 3 || len(c.OMIName) > 128 {
		errs = append(errs, fmt.Errorf("ami_name must be between 3 and 128 characters long"))
	}

	if c.OMIName != templateCleanOMIName(c.OMIName) {
		errs = append(errs, fmt.Errorf("OMIName should only contain "+
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

func (c *OMIConfig) prepareRegions(accessConfig *AccessConfig) (errs []error) {
	if len(c.OMIRegions) > 0 {
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(c.OMIRegions))

		for _, region := range c.OMIRegions {
			// If we already saw the region, then don't look again
			if _, ok := regionSet[region]; ok {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = struct{}{}

			// Make sure that if we have region_kms_key_ids defined,
			// the regions in ami_regions are also in region_kms_key_ids
			if len(c.OMIRegionKMSKeyIDs) > 0 {
				if _, ok := c.OMIRegionKMSKeyIDs[region]; !ok {
					errs = append(errs, fmt.Errorf("Region %s is in ami_regions but not in region_kms_key_ids", region))
				}
			}
			if (accessConfig != nil) && (region == accessConfig.RawRegion) {
				// make sure we don't try to copy to the region we originally
				// create the OMI in.
				log.Printf("Cannot copy OMI to AWS session region '%s', deleting it from `ami_regions`.", region)
				continue
			}
			regions = append(regions, region)
		}

		c.OMIRegions = regions
	}
	return errs
}

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html
func validateKmsKey(kmsKey string) (valid bool) {
	kmsKeyIdPattern := `[a-f0-9-]+$`
	aliasPattern := `alias/[a-zA-Z0-9:/_-]+$`
	kmsArnStartPattern := `^arn:aws:kms:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12}):`
	if regexp.MustCompile(fmt.Sprintf("^%s", kmsKeyIdPattern)).MatchString(kmsKey) {
		return true
	}
	if regexp.MustCompile(fmt.Sprintf("^%s", aliasPattern)).MatchString(kmsKey) {
		return true
	}
	if regexp.MustCompile(fmt.Sprintf("%skey/%s", kmsArnStartPattern, kmsKeyIdPattern)).MatchString(kmsKey) {
		return true
	}
	if regexp.MustCompile(fmt.Sprintf("%s%s", kmsArnStartPattern, aliasPattern)).MatchString(kmsKey) {
		return true
	}
	return false
}

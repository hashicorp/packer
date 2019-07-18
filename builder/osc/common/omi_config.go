package common

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

// OMIConfig is for common configuration related to creating OMIs.
type OMIConfig struct {
	OMIName                 string   `mapstructure:"omi_name"`
	OMIDescription          string   `mapstructure:"omi_description"`
	OMIVirtType             string   `mapstructure:"omi_virtualization_type"`
	OMIAccountIDs           []string `mapstructure:"omi_account_ids"`
	OMIGroups               []string `mapstructure:"omi_groups"`
	OMIProductCodes         []string `mapstructure:"omi_product_codes"`
	OMIRegions              []string `mapstructure:"omi_regions"`
	OMISkipRegionValidation bool     `mapstructure:"skip_region_validation"`
	OMITags                 TagMap   `mapstructure:"tags"`
	OMIENASupport           *bool    `mapstructure:"ena_support"`
	OMISriovNetSupport      bool     `mapstructure:"sriov_support"`
	OMIForceDeregister      bool     `mapstructure:"force_deregister"`
	OMIForceDeleteSnapshot  bool     `mapstructure:"force_delete_snapshot"`
	SnapshotTags            TagMap   `mapstructure:"snapshot_tags"`
	SnapshotAccountIDs      []string `mapstructure:"snapshot_account_ids"`
	SnapshotGroups          []string `mapstructure:"snapshot_groups"`
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
		errs = append(errs, fmt.Errorf("omi_name must be specified"))
	}

	errs = append(errs, c.prepareRegions(accessConfig)...)

	if len(c.OMIName) < 3 || len(c.OMIName) > 128 {
		errs = append(errs, fmt.Errorf("omi_name must be between 3 and 128 characters long"))
	}

	if c.OMIName != templateCleanResourceName(c.OMIName) {
		errs = append(errs, fmt.Errorf("OMIName should only contain "+
			"alphanumeric characters, parentheses (()), square brackets ([]), spaces "+
			"( ), periods (.), slashes (/), dashes (-), single quotes ('), at-signs "+
			"(@), or underscores(_). You can use the `clean_omi_name` template "+
			"filter to automatically clean your omi name."))
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

			if (accessConfig != nil) && (region == accessConfig.RawRegion) {
				// make sure we don't try to copy to the region we originally
				// create the OMI in.
				log.Printf("Cannot copy OMI to OUTSCALE session region '%s', deleting it from `omi_regions`.", region)
				continue
			}
			regions = append(regions, region)
		}

		c.OMIRegions = regions
	}
	return errs
}

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

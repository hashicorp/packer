//go:generate struct-markdown

package common

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

// AMIConfig is for common configuration related to creating AMIs.
type AMIConfig struct {

	// The name of the resulting AMI that will appear when
	// managing AMIs in the AWS console or via APIs. This must be unique. To help
	// make this unique, use a function like timestamp (see [template
	// engine](../templates/engine.html) for more info).
	AMIName string `mapstructure:"ami_name" required:"true"`
	// The description to set for the resulting
	// AMI(s). By default this description is empty. This is a template
	// engine, see Build template
	// data for more information.
	AMIDescription string `mapstructure:"ami_description" required:"false"`
	// The description to set for the resulting AMI(s). By default this
	// description is empty. This is a [template
	// engine](../templates/engine.html), see [Build template
	// data](#build-template-data) for more information.
	AMIVirtType string `mapstructure:"ami_virtualization_type" required:"false"`
	// A list of account IDs that have access to
	// launch the resulting AMI(s). By default no additional users other than the
	// user creating the AMI has permissions to launch it.
	AMIUsers []string `mapstructure:"ami_users" required:"false"`
	// A list of groups that have access to
	// launch the resulting AMI(s). By default no groups have permission to launch
	// the AMI. all will make the AMI publicly accessible.
	AMIGroups []string `mapstructure:"ami_groups" required:"false"`
	// A list of product codes to
	// associate with the AMI. By default no product codes are associated with the
	// AMI.
	AMIProductCodes []string `mapstructure:"ami_product_codes" required:"false"`
	// A list of regions to copy the AMI to.
	// Tags and attributes are copied along with the AMI. AMI copying takes time
	// depending on the size of the AMI, but will generally take many minutes.
	AMIRegions []string `mapstructure:"ami_regions" required:"false"`
	// Set to true if you want to skip
	// validation of the ami_regions configuration option. Default false.
	AMISkipRegionValidation bool `mapstructure:"skip_region_validation" required:"false"`
	// Tags applied to the AMI. This is a
	// [template engine](/docs/templates/engine.html), see [Build template
	// data](#build-template-data) for more information.
	AMITags TagMap `mapstructure:"tags" required:"false"`
	// Enable enhanced networking (ENA but not
	// SriovNetSupport) on HVM-compatible AMIs. If set, add
	// ec2:ModifyInstanceAttribute to your AWS IAM policy. If false, this will
	// disable enhanced networking in the final AMI as opposed to passing the
	// setting through unchanged from the source. Note: you must make sure
	// enhanced networking is enabled on your instance. [Amazon's
	// documentation on enabling enhanced
	// networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
	AMIENASupport *bool `mapstructure:"ena_support" required:"false"`
	// Enable enhanced networking (SriovNetSupport but not ENA) on
	// HVM-compatible AMIs. If true, add `ec2:ModifyInstanceAttribute` to your
	// AWS IAM policy. Note: you must make sure enhanced networking is enabled
	// on your instance. See [Amazon's documentation on enabling enhanced
	// networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
	// Default `false`.
	AMISriovNetSupport bool `mapstructure:"sriov_support" required:"false"`
	// Force Packer to first deregister an existing
	// AMI if one with the same name already exists. Default false.
	AMIForceDeregister bool `mapstructure:"force_deregister" required:"false"`
	// Force Packer to delete snapshots
	// associated with AMIs, which have been deregistered by force_deregister.
	// Default false.
	AMIForceDeleteSnapshot bool `mapstructure:"force_delete_snapshot" required:"false"`
	// Whether or not to encrypt the resulting AMI when
	// copying a provisioned instance to an AMI. By default, Packer will keep the
	// encryption setting to what it was in the source image. Setting false will
	// result in an unencrypted image, and true will result in an encrypted one.
	AMIEncryptBootVolume *bool `mapstructure:"encrypt_boot" required:"false"`
	// ID, alias or ARN of the KMS key to use for boot volume encryption. This
	// only applies to the main `region`, other regions where the AMI will be
	// copied will be encrypted by the default EBS KMS key. For valid formats
	// see *KmsKeyId* in the [AWS API docs -
	// CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
	// This field is validated by Packer, when using an alias, you will have to
	// prefix `kms_key_id` with `alias/`.
	AMIKmsKeyId string `mapstructure:"kms_key_id" required:"false"`
	// regions to copy the ami to, along with the custom kms key id (alias or
	// arn) to use for encryption for that region. Keys must match the regions
	// provided in `ami_regions`. If you just want to encrypt using a default
	// ID, you can stick with `kms_key_id` and `ami_regions`. If you want a
	// region to be encrypted with that region's default key ID, you can use an
	// empty string `""` instead of a key id in this map. (e.g. `"us-east-1":
	// ""`) However, you cannot use default key IDs if you are using this in
	// conjunction with `snapshot_users` -- in that situation you must use
	// custom keys. For valid formats see *KmsKeyId* in the [AWS API docs -
	// CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
	//
	// This option supercedes the `kms_key_id` option -- if you set both, and
	// they are different, Packer will respect the value in
	// `region_kms_key_ids` for your build region and silently disregard the
	// value provided in `kms_key_id`.
	AMIRegionKMSKeyIDs map[string]string `mapstructure:"region_kms_key_ids" required:"false"`
	// If true, Packer will not check whether an AMI with the `ami_name` exists
	// in the region it is building in. It will use an intermediary AMI name,
	// which it will not convert to an AMI in the build region. It will copy
	// the intermediary AMI into any regions provided in `ami_regions`, then
	// delete the intermediary AMI. Default `false`.
	AMISkipBuildRegion bool `mapstructure:"skip_save_build_region"`
	// Tags to apply to snapshot.
	// They will override AMI tags if already applied to snapshot. This is a
	// [template engine](../templates/engine.html), see [Build template
	// data](#build-template-data) for more information.
	SnapshotTags TagMap `mapstructure:"snapshot_tags" required:"false"`
	// A list of account IDs that have
	// access to create volumes from the snapshot(s). By default no additional
	// users other than the user creating the AMI has permissions to create
	// volumes from the backing snapshot(s).
	SnapshotUsers []string `mapstructure:"snapshot_users" required:"false"`
	// A list of groups that have access to
	// create volumes from the snapshot(s). By default no groups have permission
	// to create volumes from the snapshot(s). all will make the snapshot
	// publicly accessible.
	SnapshotGroups []string `mapstructure:"snapshot_groups" required:"false"`
}

func stringInSlice(s []string, searchstr string) bool {
	for _, item := range s {
		if item == searchstr {
			return true
		}
	}
	return false
}

func (c *AMIConfig) Prepare(accessConfig *AccessConfig, ctx *interpolate.Context) []error {
	var errs []error

	if c.AMIName == "" {
		errs = append(errs, fmt.Errorf("ami_name must be specified"))
	}

	// Make sure that if we have region_kms_key_ids defined,
	//  the regions in region_kms_key_ids are also in ami_regions
	if len(c.AMIRegionKMSKeyIDs) > 0 {
		for kmsKeyRegion := range c.AMIRegionKMSKeyIDs {
			if !stringInSlice(c.AMIRegions, kmsKeyRegion) {
				errs = append(errs, fmt.Errorf("Region %s is in region_kms_key_ids but not in ami_regions", kmsKeyRegion))
			}
		}
	}

	errs = append(errs, c.prepareRegions(accessConfig)...)

	// Prevent sharing of default KMS key encrypted volumes with other aws users
	if len(c.AMIUsers) > 0 {
		if len(c.AMIKmsKeyId) == 0 && c.AMIEncryptBootVolume != nil && *c.AMIEncryptBootVolume {
			errs = append(errs, fmt.Errorf("Cannot share AMI encrypted with default KMS key"))
		}
		if len(c.AMIRegionKMSKeyIDs) > 0 {
			for _, kmsKey := range c.AMIRegionKMSKeyIDs {
				if len(kmsKey) == 0 {
					errs = append(errs, fmt.Errorf("Cannot share AMI encrypted with default KMS key for other regions"))
				}
			}
		}
	}

	var kmsKeys []string
	if len(c.AMIKmsKeyId) > 0 {
		kmsKeys = append(kmsKeys, c.AMIKmsKeyId)
	}
	if len(c.AMIRegionKMSKeyIDs) > 0 {
		for _, kmsKey := range c.AMIRegionKMSKeyIDs {
			if len(kmsKey) == 0 {
				kmsKeys = append(kmsKeys, c.AMIKmsKeyId)
			}
		}
	}
	for _, kmsKey := range kmsKeys {
		if !validateKmsKey(kmsKey) {
			errs = append(errs, fmt.Errorf("%s is not a valid KMS Key Id.", kmsKey))
		}
	}

	if len(c.SnapshotUsers) > 0 {
		if len(c.AMIKmsKeyId) == 0 && c.AMIEncryptBootVolume != nil && *c.AMIEncryptBootVolume {
			errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
		}
		if len(c.AMIRegionKMSKeyIDs) > 0 {
			for _, kmsKey := range c.AMIRegionKMSKeyIDs {
				if len(kmsKey) == 0 {
					errs = append(errs, fmt.Errorf("Cannot share snapshot encrypted with default KMS key"))
				}
			}
		}
	}

	if len(c.AMIName) < 3 || len(c.AMIName) > 128 {
		errs = append(errs, fmt.Errorf("ami_name must be between 3 and 128 characters long"))
	}

	if c.AMIName != templateCleanAMIName(c.AMIName) {
		errs = append(errs, fmt.Errorf("AMIName should only contain "+
			"alphanumeric characters, parentheses (()), square brackets ([]), spaces "+
			"( ), periods (.), slashes (/), dashes (-), single quotes ('), at-signs "+
			"(@), or underscores(_). You can use the `clean_resource_name` template "+
			"filter to automatically clean your ami name."))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AMIConfig) prepareRegions(accessConfig *AccessConfig) (errs []error) {
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

			// Make sure that if we have region_kms_key_ids defined,
			// the regions in ami_regions are also in region_kms_key_ids
			if len(c.AMIRegionKMSKeyIDs) > 0 {
				if _, ok := c.AMIRegionKMSKeyIDs[region]; !ok {
					errs = append(errs, fmt.Errorf("Region %s is in ami_regions but not in region_kms_key_ids", region))
				}
			}
			if (accessConfig != nil) && (region == accessConfig.RawRegion) {
				// make sure we don't try to copy to the region we originally
				// create the AMI in.
				log.Printf("Cannot copy AMI to AWS session region '%s', deleting it from `ami_regions`.", region)
				continue
			}
			regions = append(regions, region)
		}

		c.AMIRegions = regions
	}
	return errs
}

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html
func validateKmsKey(kmsKey string) (valid bool) {
	kmsKeyIdPattern := `[a-f0-9-]+$`
	aliasPattern := `alias/[a-zA-Z0-9:/_-]+$`
	kmsArnStartPattern := `^arn:aws(-us-gov)?:kms:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12}):`
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

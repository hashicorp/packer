package cvm

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

type TencentCloudImageConfig struct {
	// The name you want to create your customize image,
    // it should be composed of no more than 20 characters, of letters, numbers
    // or minus sign.
	ImageName          string   `mapstructure:"image_name" required:"true"`
	// Image description.
	ImageDescription   string   `mapstructure:"image_description" required:"false"`
	// Whether shutdown cvm to create Image. Default value is
    // false.
	Reboot             bool     `mapstructure:"reboot" required:"false"`
	// Whether to force power off cvm when create image.
    // Default value is false.
	ForcePoweroff      bool     `mapstructure:"force_poweroff" required:"false"`
	// Whether enable Sysprep during creating windows image.
	Sysprep            bool     `mapstructure:"sysprep" required:"false"`
	ImageForceDelete   bool     `mapstructure:"image_force_delete"`
	// regions that will be copied to after
    // your image created.
	ImageCopyRegions   []string `mapstructure:"image_copy_regions" required:"false"`
	// accounts that will be shared to
    // after your image created.
	ImageShareAccounts []string `mapstructure:"image_share_accounts" required:"false"`
	// Do not check region and zone when validate.
	SkipValidation     bool     `mapstructure:"skip_region_validation" required:"false"`
}

func (cf *TencentCloudImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	cf.ForcePoweroff = true
	if cf.ImageName == "" {
		errs = append(errs, fmt.Errorf("image_name must be set"))
	} else if len(cf.ImageName) > 20 {
		errs = append(errs, fmt.Errorf("image_num length should not exceed 20 characters"))
	} else {
		regex := regexp.MustCompile("^[0-9a-zA-Z\\-]+$")
		if !regex.MatchString(cf.ImageName) {
			errs = append(errs, fmt.Errorf("image_name can only be composed of letters, numbers and minus sign"))
		}
	}

	if len(cf.ImageDescription) > 60 {
		errs = append(errs, fmt.Errorf("image_description length should not exceed 60 characters"))
	}

	if len(cf.ImageCopyRegions) > 0 {
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(cf.ImageCopyRegions))

		for _, region := range cf.ImageCopyRegions {
			if _, ok := regionSet[region]; ok {
				continue
			}

			regionSet[region] = struct{}{}

			if !cf.SkipValidation {
				if err := validRegion(region); err != nil {
					errs = append(errs, err)
					continue
				}
			}
			regions = append(regions, region)
		}
		cf.ImageCopyRegions = regions
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validRegion(region string) error {
	for _, valid := range ValidRegions {
		if Region(region) == valid {
			return nil
		}
	}
	return fmt.Errorf("unknown region: %s", region)
}

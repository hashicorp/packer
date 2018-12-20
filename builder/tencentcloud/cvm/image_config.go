package cvm

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/packer/template/interpolate"
)

type TencentCloudImageConfig struct {
	ImageName          string   `mapstructure:"image_name"`
	ImageDescription   string   `mapstructure:"image_description"`
	Reboot             bool     `mapstructure:"reboot"`
	ForcePoweroff      bool     `mapstructure:"force_poweroff"`
	Sysprep            bool     `mapstructure:"sysprep"`
	ImageForceDelete   bool     `mapstructure:"image_force_delete"`
	ImageCopyRegions   []string `mapstructure:"image_copy_regions"`
	ImageShareAccounts []string `mapstructure:"image_share_accounts"`
	SkipValidation     bool     `mapstructure:"skip_region_validation"`
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

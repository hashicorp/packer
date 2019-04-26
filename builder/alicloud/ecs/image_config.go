package ecs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

type AlicloudDiskDevice struct {
	DiskName           string `mapstructure:"disk_name"`
	DiskCategory       string `mapstructure:"disk_category"`
	DiskSize           int    `mapstructure:"disk_size"`
	SnapshotId         string `mapstructure:"disk_snapshot_id"`
	Description        string `mapstructure:"disk_description"`
	DeleteWithInstance bool   `mapstructure:"disk_delete_with_instance"`
	Device             string `mapstructure:"disk_device"`
	Encrypted          *bool  `mapstructure:"disk_encrypted"`
}

type AlicloudDiskDevices struct {
	ECSSystemDiskMapping  AlicloudDiskDevice   `mapstructure:"system_disk_mapping"`
	ECSImagesDiskMappings []AlicloudDiskDevice `mapstructure:"image_disk_mappings"`
}

type AlicloudImageConfig struct {
	AlicloudImageName                 string            `mapstructure:"image_name"`
	AlicloudImageVersion              string            `mapstructure:"image_version"`
	AlicloudImageDescription          string            `mapstructure:"image_description"`
	AlicloudImageShareAccounts        []string          `mapstructure:"image_share_account"`
	AlicloudImageUNShareAccounts      []string          `mapstructure:"image_unshare_account"`
	AlicloudImageDestinationRegions   []string          `mapstructure:"image_copy_regions"`
	AlicloudImageDestinationNames     []string          `mapstructure:"image_copy_names"`
	ImageEncrypted                    *bool             `mapstructure:"image_encrypted"`
	AlicloudImageForceDelete          bool              `mapstructure:"image_force_delete"`
	AlicloudImageForceDeleteSnapshots bool              `mapstructure:"image_force_delete_snapshots"`
	AlicloudImageForceDeleteInstances bool              `mapstructure:"image_force_delete_instances"`
	AlicloudImageIgnoreDataDisks      bool              `mapstructure:"image_ignore_data_disks"`
	AlicloudImageSkipRegionValidation bool              `mapstructure:"skip_region_validation"`
	AlicloudImageTags                 map[string]string `mapstructure:"tags"`
	AlicloudDiskDevices               `mapstructure:",squash"`
}

func (c *AlicloudImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.AlicloudImageName == "" {
		errs = append(errs, fmt.Errorf("image_name must be specified"))
	} else if len(c.AlicloudImageName) < 2 || len(c.AlicloudImageName) > 128 {
		errs = append(errs, fmt.Errorf("image_name must less than 128 letters and more than 1 letters"))
	} else if strings.HasPrefix(c.AlicloudImageName, "http://") ||
		strings.HasPrefix(c.AlicloudImageName, "https://") {
		errs = append(errs, fmt.Errorf("image_name can't start with 'http://' or 'https://'"))
	}
	reg := regexp.MustCompile("\\s+")
	if reg.FindString(c.AlicloudImageName) != "" {
		errs = append(errs, fmt.Errorf("image_name can't include spaces"))
	}

	if len(c.AlicloudImageDestinationRegions) > 0 {
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(c.AlicloudImageDestinationRegions))

		for _, region := range c.AlicloudImageDestinationRegions {
			// If we already saw the region, then don't look again
			if _, ok := regionSet[region]; ok {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = struct{}{}
			regions = append(regions, region)
		}

		c.AlicloudImageDestinationRegions = regions
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

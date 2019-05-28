package ecs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

type AlicloudDiskDevice struct {
	// The value of disk name is blank by default. [2,
    // 128] English or Chinese characters, must begin with an
    // uppercase/lowercase letter or Chinese character. Can contain numbers,
    // ., _ and -. The disk name will appear on the console. It cannot
    // begin with http:// or https://.
	DiskName           string `mapstructure:"disk_name" required:"false"`
	// Category of the system disk. Optional values
    // are:
    //     -   cloud - general cloud disk
    //     -   cloud_efficiency - efficiency cloud disk
    //     -   cloud_ssd - cloud SSD
	DiskCategory       string `mapstructure:"disk_category" required:"false"`
	// Size of the system disk, measured in GiB. Value
    // range: [20, 500]. The specified value must be equal to or greater
    // than max{20, ImageSize}. Default value: max{40, ImageSize}.
	DiskSize           int    `mapstructure:"disk_size" required:"false"`
	// Snapshots are used to create the data
    // disk After this parameter is specified, Size is ignored. The actual
    // size of the created disk is the size of the specified snapshot.
	SnapshotId         string `mapstructure:"disk_snapshot_id" required:"false"`
	// The value of disk description is blank by
    // default. [2, 256] characters. The disk description will appear on the
    // console. It cannot begin with http:// or https://.
	Description        string `mapstructure:"disk_description" required:"false"`
	// Whether or not the disk is
    // released along with the instance:
	DeleteWithInstance bool   `mapstructure:"disk_delete_with_instance" required:"false"`
	// Device information of the related instance:
    // such as /dev/xvdb It is null unless the Status is In_use.
	Device             string `mapstructure:"disk_device" required:"false"`
	// Whether or not to encrypt the data disk. 
    // If this option is set to true, the data disk will be encryped and corresponding snapshot in the target image will also be encrypted. By 
    // default, if this is an extra data disk, Packer will not encrypt the 
    // data disk. Otherwise, Packer will keep the encryption setting to what 
    // it was in the source image. Please refer to Introduction of ECS disk encryption 
    // for more details.
	Encrypted          *bool  `mapstructure:"disk_encrypted" required:"false"`
}

type AlicloudDiskDevices struct {
	// Image disk mapping for system
    // disk.
	ECSSystemDiskMapping  AlicloudDiskDevice   `mapstructure:"system_disk_mapping" required:"false"`
	// Add one or more data
    // disks to the image.
	ECSImagesDiskMappings []AlicloudDiskDevice `mapstructure:"image_disk_mappings" required:"false"`
}

type AlicloudImageConfig struct {
	// The name of the user-defined image, [2, 128]
    // English or Chinese characters. It must begin with an uppercase/lowercase
    // letter or a Chinese character, and may contain numbers, _ or -. It
    // cannot begin with http:// or https://.
	AlicloudImageName                 string            `mapstructure:"image_name" required:"true"`
	// The version number of the image, with a length
    // limit of 1 to 40 English characters.
	AlicloudImageVersion              string            `mapstructure:"image_version" required:"false"`
	// The description of the image, with a length
    // limit of 0 to 256 characters. Leaving it blank means null, which is the
    // default value. It cannot begin with http:// or https://.
	AlicloudImageDescription          string            `mapstructure:"image_description" required:"false"`
	// The IDs of to-be-added Aliyun
    // accounts to which the image is shared. The number of accounts is 1 to 10.
    // If number of accounts is greater than 10, this parameter is ignored.
	AlicloudImageShareAccounts        []string          `mapstructure:"image_share_account" required:"false"`
	AlicloudImageUNShareAccounts      []string          `mapstructure:"image_unshare_account"`
	// Copy to the destination regionIds.
	AlicloudImageDestinationRegions   []string          `mapstructure:"image_copy_regions" required:"false"`
	// The name of the destination image,
    // [2, 128] English or Chinese characters. It must begin with an
    // uppercase/lowercase letter or a Chinese character, and may contain numbers,
    // _ or -. It cannot begin with http:// or https://.
	AlicloudImageDestinationNames     []string          `mapstructure:"image_copy_names" required:"false"`
	// Whether or not to encrypt the target images,            including those copied if image_copy_regions is specified. If this option 
    // is set to true, a temporary image will be created from the provisioned 
    // instance in the main region and an encrypted copy will be generated in the 
    // same region. By default, Packer will keep the encryption setting to what 
    // it was in the source image.
	ImageEncrypted                    *bool             `mapstructure:"image_encrypted" required:"false"`
	// If this value is true, when the target
    // image names including those copied are duplicated with existing images, it
    // will delete the existing images and then create the target images,
    // otherwise, the creation will fail. The default value is false. Check
    // image_name and image_copy_names options for names of target images. If
    // -force option is
    // provided in build command, this option can be omitted and taken as true.
	AlicloudImageForceDelete          bool              `mapstructure:"image_force_delete" required:"false"`
	// If this value is true, when
    // delete the duplicated existing images, the source snapshots of those images
    // will be delete either. If
    // -force option is
    // provided in build command, this option can be omitted and taken as true.
	AlicloudImageForceDeleteSnapshots bool              `mapstructure:"image_force_delete_snapshots" required:"false"`
	AlicloudImageForceDeleteInstances bool              `mapstructure:"image_force_delete_instances"`
	// If this value is true, the image
    // created will not include any snapshot of data disks. This option would be
    // useful for any circumstance that default data disks with instance types are
    // not concerned. The default value is false.
	AlicloudImageIgnoreDataDisks      bool              `mapstructure:"image_ignore_data_disks" required:"false"`
	// The region validation can be skipped
    // if this value is true, the default value is false.
	AlicloudImageSkipRegionValidation bool              `mapstructure:"skip_region_validation" required:"false"`
	// Tags applied to the destination
    // image and relevant snapshots.
	AlicloudImageTags                 map[string]string `mapstructure:"tags" required:"false"`
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

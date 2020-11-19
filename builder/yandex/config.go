//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package yandex

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

const defaultGpuPlatformID = "gpu-standard-v1"
const defaultPlatformID = "standard-v1"
const defaultMaxRetries = 3
const defaultZone = "ru-central1-a"

var reImageFamily = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Communicator        communicator.Config `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`

	// The folder ID that will be used to launch instances and store images.
	// Alternatively you may set value by environment variable `YC_FOLDER_ID`.
	// To use a different folder for looking up the source image or saving the target image to
	// check options 'source_image_folder_id' and 'target_image_folder_id'.
	FolderID string `mapstructure:"folder_id" required:"true"`
	// Service account identifier to assign to instance.
	ServiceAccountID string `mapstructure:"service_account_id" required:"false"`
	// The name of the disk, if unset the instance name
	// will be used.
	DiskName string `mapstructure:"disk_name" required:"false"`
	// The size of the disk in GB. This defaults to `10`, which is 10GB.
	DiskSizeGb int `mapstructure:"disk_size_gb" required:"false"`
	// Specify disk type for the launched instance. Defaults to `network-hdd`.
	DiskType string `mapstructure:"disk_type" required:"false"`
	// Key/value pair labels to apply to the disk.
	DiskLabels map[string]string `mapstructure:"disk_labels" required:"false"`
	// The description of the resulting image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	//  The family name of the resulting image.
	ImageFamily string `mapstructure:"image_family" required:"false"`
	// Key/value pair labels to apply to the created image.
	ImageLabels map[string]string `mapstructure:"image_labels" required:"false"`
	// Minimum size of the disk that will be created from built image, specified in gigabytes.
	// Should be more or equal to `disk_size_gb`.
	ImageMinDiskSizeGb int `mapstructure:"image_min_disk_size_gb" required:"false"`
	// The unique name of the resulting image. Defaults to
	// `packer-{{timestamp}}`.
	ImageName string `mapstructure:"image_name" required:"false"`
	// License IDs that indicate which licenses are attached to resulting image.
	ImageProductIDs []string `mapstructure:"image_product_ids" required:"false"`
	// The number of cores available to the instance.
	InstanceCores int `mapstructure:"instance_cores" required:"false"`
	// The number of GPU available to the instance.
	InstanceGpus int `mapstructure:"instance_gpus"`
	// The amount of memory available to the instance, specified in gigabytes.
	InstanceMemory int `mapstructure:"instance_mem_gb" required:"false"`
	// The name assigned to the instance.
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// Key/value pair labels to apply to the launched instance.
	Labels map[string]string `mapstructure:"labels" required:"false"`
	// Identifier of the hardware platform configuration for the instance. This defaults to `standard-v1`.
	PlatformID string `mapstructure:"platform_id" required:"false"`
	// Metadata applied to the launched instance.
	Metadata map[string]string `mapstructure:"metadata" required:"false"`
	// Metadata applied to the launched instance.
	// The values in this map are the paths to the content files for the corresponding metadata keys.
	MetadataFromFile map[string]string `mapstructure:"metadata_from_file"`
	// Launch a preemptible instance. This defaults to `false`.
	Preemptible bool `mapstructure:"preemptible"`
	// File path to save serial port output of the launched instance.
	SerialLogFile string `mapstructure:"serial_log_file" required:"false"`
	// The source image family to create the new image
	// from. You can also specify source_image_id instead. Just one of a source_image_id or
	// source_image_family must be specified. Example: `ubuntu-1804-lts`.
	SourceImageFamily string `mapstructure:"source_image_family" required:"true"`
	// The ID of the folder containing the source image.
	SourceImageFolderID string `mapstructure:"source_image_folder_id" required:"false"`
	// The source image ID to use to create the new image from.
	SourceImageID string `mapstructure:"source_image_id" required:"false"`
	// The source image name to use to create the new image
	// from. Name will be looked up in `source_image_folder_id`.
	SourceImageName string `mapstructure:"source_image_name"`
	// The Yandex VPC subnet id to use for
	// the launched instance. Note, the zone of the subnet must match the
	// zone in which the VM is launched.
	SubnetID string `mapstructure:"subnet_id" required:"false"`
	// The ID of the folder to save built image in.
	// This defaults to value of 'folder_id'.
	TargetImageFolderID string `mapstructure:"target_image_folder_id" required:"false"`
	// If set to true, then launched instance will have external internet
	// access.
	UseIPv4Nat bool `mapstructure:"use_ipv4_nat" required:"false"`
	// Set to true to enable IPv6 for the instance being
	// created. This defaults to `false`, or not enabled.
	//
	// -> **Note**: Usage of IPv6 will be available in the future.
	UseIPv6 bool `mapstructure:"use_ipv6" required:"false"`
	// If true, use the instance's internal IP address
	// instead of its external IP during building.
	UseInternalIP bool `mapstructure:"use_internal_ip" required:"false"`
	// The name of the zone to launch the instance.  This defaults to `ru-central1-a`.
	Zone string `mapstructure:"zone" required:"false"`

	ctx interpolate.Context
	// The time to wait for instance state changes.
	// Defaults to `5m`.
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, c.AccessConfig.Prepare(&c.ctx)...)

	if c.SerialLogFile != "" {
		if _, err := os.Stat(c.SerialLogFile); os.IsExist(err) {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Serial log file %s already exist", c.SerialLogFile))
		}
	}

	if c.InstanceCores == 0 {
		c.InstanceCores = 2
	}

	if c.InstanceMemory == 0 {
		c.InstanceMemory = 4
	}

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 10
	}

	if c.DiskType == "" {
		c.DiskType = "network-hdd"
	}

	if c.ImageMinDiskSizeGb == 0 {
		c.ImageMinDiskSizeGb = c.DiskSizeGb
	}

	if c.ImageMinDiskSizeGb < c.DiskSizeGb {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("Invalid image_min_disk_size value (%d): Must be equal or greate than disk_size_gb (%d)",
				c.ImageMinDiskSizeGb, c.DiskSizeGb))
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Unable to render default image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	if len(c.ImageFamily) > 63 {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Invalid image family: Must not be longer than 63 characters"))
	}

	if c.ImageFamily != "" {
		if !reImageFamily.MatchString(c.ImageFamily) {
			errs = packersdk.MultiErrorAppend(errs,
				errors.New("Invalid image family: The first character must be a "+
					"lowercase letter, and all following characters must be a dash, "+
					"lowercase letter, or digit, except the last character, which cannot be a dash"))
		}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName + "-disk"
	}

	if c.PlatformID == "" {
		c.PlatformID = defaultPlatformID
		if c.InstanceGpus != 0 {
			c.PlatformID = defaultGpuPlatformID
		}
	}

	if es := c.Communicator.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.
	if c.SourceImageID == "" {
		if c.SourceImageFamily == "" && c.SourceImageName == "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("a source_image_name or source_image_family must be specified"))
		}
		if c.SourceImageFamily != "" && c.SourceImageName != "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("one of source_image_name or source_image_family must be specified, not both"))
		}
	}

	if c.Zone == "" {
		c.Zone = defaultZone
	}

	if c.MaxRetries == 0 {
		c.MaxRetries = defaultMaxRetries
	}

	if c.FolderID == "" {
		c.FolderID = os.Getenv("YC_FOLDER_ID")
	}

	if c.FolderID == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("a folder_id must be specified"))
	}

	if c.TargetImageFolderID == "" {
		c.TargetImageFolderID = c.FolderID
	}

	for key, file := range c.MetadataFromFile {
		if _, err := os.Stat(file); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("cannot access file '%s' with content for value of metadata key '%s': %s", file, key, err))
		}
	}

	if c.StateTimeout == 0 {
		c.StateTimeout = 5 * time.Minute
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

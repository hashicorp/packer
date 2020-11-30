//go:generate struct-markdown

package yandex

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

const (
	defaultPlatformID    = "standard-v2"
	defaultZone          = "ru-central1-a"
	defaultGpuPlatformID = "gpu-standard-v1"
)

var reImageFamily = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

type CommonConfig struct {

	// File path to save serial port output of the launched instance.
	SerialLogFile string `mapstructure:"serial_log_file" required:"false"`

	InstanceConfig `mapstructure:",squash"`
	DiskConfig     `mapstructure:",squash"`
	NetworkConfig  `mapstructure:",squash"`
	CloudConfig    `mapstructure:",squash"`
}

func (c *CommonConfig) Prepare(errs *packer.MultiError) *packer.MultiError {

	if c.SerialLogFile != "" {
		if _, err := os.Stat(c.SerialLogFile); os.IsExist(err) {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Serial log file %s already exist", c.SerialLogFile))
		}
	}

	errs = c.CloudConfig.Prepare(errs)
	errs = c.InstanceConfig.Prepare(errs)
	errs = c.DiskConfig.Prepare(errs)
	errs = c.NetworkConfig.Prepare(errs)

	return errs
}

type CloudConfig struct {
	// The folder ID that will be used to launch instances and store images.
	// Alternatively you may set value by environment variable `YC_FOLDER_ID`.
	// To use a different folder for looking up the source image or saving the target image to
	// check options 'source_image_folder_id' and 'target_image_folder_id'.
	FolderID string `mapstructure:"folder_id" required:"true"`
}

func (c *CloudConfig) Prepare(errs *packer.MultiError) *packer.MultiError {
	if c.FolderID == "" {
		c.FolderID = os.Getenv("YC_FOLDER_ID")
	}

	if c.FolderID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a folder_id must be specified"))
	}
	return errs
}

type DiskConfig struct {
	// The name of the disk, if unset the instance name
	// will be used.
	DiskName string `mapstructure:"disk_name" required:"false"`
	// The size of the disk in GB. This defaults to 10/100GB.
	DiskSizeGb int `mapstructure:"disk_size_gb" required:"false"`
	// Specify disk type for the launched instance. Defaults to `network-ssd`.
	DiskType string `mapstructure:"disk_type" required:"false"`
	// Key/value pair labels to apply to the disk.
	DiskLabels map[string]string `mapstructure:"disk_labels" required:"false"`
}

func (c *DiskConfig) Prepare(errs *packer.MultiError) *packer.MultiError {

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 10
	}

	if c.DiskType == "" {
		c.DiskType = "network-ssd"
	}

	return errs
}

type NetworkConfig struct {
	// The Yandex VPC subnet id to use for
	// the launched instance. Note, the zone of the subnet must match the
	// zone in which the VM is launched.
	SubnetID string `mapstructure:"subnet_id" required:"false"`
	// The name of the zone to launch the instance.  This defaults to `ru-central1-a`.
	Zone string `mapstructure:"zone" required:"false"`

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
}

func (c *NetworkConfig) Prepare(errs *packer.MultiError) *packer.MultiError {
	if c.Zone == "" {
		c.Zone = defaultZone
	}

	// if c.UseIPv4Nat && c.UseIPv6 {
	// 	errs = packer.MultiErrorAppend(
	// 		errors.New("one of use_ipv4_nat or use_ipv6 key file must be specified, not both"),
	// 		errs,
	// 	)
	// }
	return errs
}

type ImageConfig struct {
	// The name of the resulting image, which contains 1-63 characters and only
	// supports lowercase English characters, numbers and hyphen. Defaults to
	// `packer-{{timestamp}}`.
	ImageName string `mapstructure:"image_name" required:"false"`
	// The description of the image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	// The family name of the image.
	ImageFamily string `mapstructure:"image_family" required:"false"`
	// Key/value pair labels to apply to the image.
	ImageLabels map[string]string `mapstructure:"image_labels" required:"false"`
	// Minimum size of the disk that will be created from built image, specified in gigabytes.
	// Should be more or equal to `disk_size_gb`.
	ImageMinDiskSizeGb int `mapstructure:"image_min_disk_size_gb" required:"false"`
	// License IDs that indicate which licenses are attached to resulting image.
	ImageProductIDs []string `mapstructure:"image_product_ids" required:"false"`
}

func (c *ImageConfig) Prepare(errs *packer.MultiError) *packer.MultiError {

	if len(c.ImageFamily) > 63 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Invalid image family: Must not be longer than 63 characters"))
	}

	if c.ImageFamily != "" {
		if !reImageFamily.MatchString(c.ImageFamily) {
			errs = packer.MultiErrorAppend(errs,
				errors.New("Invalid image family: The first character must be a "+
					"lowercase letter, and all following characters must be a dash, "+
					"lowercase letter, or digit, except the last character, which cannot be a dash"))
		}
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to render default image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	return errs
}

type InstanceConfig struct {
	// The number of cores available to the instance.
	InstanceCores int `mapstructure:"instance_cores" required:"false"`
	// The number of GPU available to the instance.
	InstanceGpus int `mapstructure:"instance_gpus" required:"false"`
	// The amount of memory available to the instance, specified in gigabytes.
	InstanceMemory int `mapstructure:"instance_mem_gb" required:"false"`
	// The name assigned to the instance.
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// Identifier of the hardware platform configuration for the instance. This defaults to `standard-v2`.
	PlatformID string `mapstructure:"platform_id" required:"false"`
	// Key/value pair labels to apply to the launched instance.
	Labels map[string]string `mapstructure:"labels" required:"false"`
	// Metadata applied to the launched instance.
	Metadata map[string]string `mapstructure:"metadata" required:"false"`
	// Metadata applied to the launched instance.
	// The values in this map are the paths to the content files for the corresponding metadata keys.
	MetadataFromFile map[string]string `mapstructure:"metadata_from_file"`
	// Launch a preemptible instance. This defaults to `false`.
	Preemptible bool `mapstructure:"preemptible"`
}

func (c *InstanceConfig) Prepare(errs *packer.MultiError) *packer.MultiError {
	if c.InstanceCores == 0 {
		c.InstanceCores = 2
	}

	if c.InstanceMemory == 0 {
		c.InstanceMemory = 4
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	for key, file := range c.MetadataFromFile {
		if _, err := os.Stat(file); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("cannot access file '%s' with content for value of metadata key '%s': %s", file, key, err))
		}
	}

	if c.PlatformID == "" {
		c.PlatformID = defaultPlatformID
		if c.InstanceGpus != 0 {
			c.PlatformID = defaultGpuPlatformID
		}
	}
	return errs
}
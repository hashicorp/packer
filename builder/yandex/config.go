package yandex

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	"github.com/yandex-cloud/go-sdk/iamkey"
)

const defaultEndpoint = "api.cloud.yandex.net:443"
const defaultZone = "ru-central1-a"

var reImageFamily = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Communicator        communicator.Config `mapstructure:",squash"`
	// Non standard api endpoint URL.
	Endpoint              string `mapstructure:"endpoint" required:"false"`
	// The folder ID that will be used to launch instances and store images.
    // Alternatively you may set value by environment variable YC_FOLDER_ID.
	FolderID              string `mapstructure:"folder_id" required:"true"`
	// Path to file with Service Account key in json format. This 
    // is an alternative method to authenticate to Yandex.Cloud. Alternatively you may set environment variable
    // YC_SERVICE_ACCOUNT_KEY_FILE.
	ServiceAccountKeyFile string `mapstructure:"service_account_key_file" required:"false"`
	// OAuth token to use to authenticate to Yandex.Cloud. Alternatively you may set
    // value by environment variable YC_TOKEN.
	Token                 string `mapstructure:"token" required:"true"`
	// The name of the disk, if unset the instance name
    // will be used.
	DiskName            string            `mapstructure:"disk_name" required:"false"`
	// The size of the disk in GB. This defaults to 10, which is 10GB.
	DiskSizeGb          int               `mapstructure:"disk_size_gb" required:"false"`
	// Specify disk type for the launched instance. Defaults to network-hdd.
	DiskType            string            `mapstructure:"disk_type" required:"false"`
	// The description of the resulting image.
	ImageDescription    string            `mapstructure:"image_description" required:"false"`
	//  The family name of the resulting image.
	ImageFamily         string            `mapstructure:"image_family" required:"false"`
	// Key/value pair labels to
    // apply to the created image.
	ImageLabels         map[string]string `mapstructure:"image_labels" required:"false"`
	// The unique name of the resulting image. Defaults to
    // packer-{{timestamp}}.
	ImageName           string            `mapstructure:"image_name" required:"false"`
	// License IDs that indicate which licenses are attached to resulting image.
	ImageProductIDs     []string          `mapstructure:"image_product_ids" required:"false"`
	// The number of cores available to the instance.
	InstanceCores       int               `mapstructure:"instance_cores" required:"false"`
	// The amount of memory available to the instance, specified in gigabytes.
	InstanceMemory      int               `mapstructure:"instance_mem_gb" required:"false"`
	// The name assigned to the instance.
	InstanceName        string            `mapstructure:"instance_name" required:"false"`
	// Key/value pair labels to apply to
    // the launched instance.
	Labels              map[string]string `mapstructure:"labels" required:"false"`
	// Identifier of the hardware platform configuration for the instance. This defaults to standard-v1.
	PlatformID          string            `mapstructure:"platform_id" required:"false"`
	// Metadata applied to the launched
    // instance.
	Metadata            map[string]string `mapstructure:"metadata" required:"false"`
	// File path to save serial port output of the launched instance.
	SerialLogFile       string            `mapstructure:"serial_log_file" required:"false"`
	// The source image family to create the new image
    // from. You can also specify source_image_id instead. Just one of a source_image_id or 
    // source_image_family must be specified. Example: ubuntu-1804-lts
	SourceImageFamily   string            `mapstructure:"source_image_family" required:"true"`
	// The ID of the folder containing the source image.
	SourceImageFolderID string            `mapstructure:"source_image_folder_id" required:"false"`
	// The source image ID to use to create the new image
    // from.
	SourceImageID       string            `mapstructure:"source_image_id" required:"false"`
	// The Yandex VPC subnet id to use for 
    // the launched instance. Note, the zone of the subnet must match the
    // zone in which the VM is launched.
	SubnetID            string            `mapstructure:"subnet_id" required:"false"`
	// If set to true, then launched instance will have external internet 
    // access.
	UseIPv4Nat          bool              `mapstructure:"use_ipv4_nat" required:"false"`
	// Set to true to enable IPv6 for the instance being
    // created. This defaults to false, or not enabled.
    // -> Note: ~> Usage of IPv6 will be available in the future.
	UseIPv6             bool              `mapstructure:"use_ipv6" required:"false"`
	// If true, use the instance's internal IP address
    // instead of its external IP during building.
	UseInternalIP       bool              `mapstructure:"use_internal_ip" required:"false"`
	// The name of the zone to launch the instance.  This defaults to ru-central1-a.
	Zone                string            `mapstructure:"zone" required:"false"`

	ctx          interpolate.Context
	// The time to wait for instance state changes.
    // Defaults to 5m.
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := &Config{}
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	if c.SerialLogFile != "" {
		if _, err := os.Stat(c.SerialLogFile); os.IsExist(err) {
			errs = packer.MultiErrorAppend(errs,
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

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName + "-disk"
	}

	if c.PlatformID == "" {
		c.PlatformID = "standard-v1"
	}

	if es := c.Communicator.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.
	if c.SourceImageID == "" && c.SourceImageFamily == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image_id or source_image_family must be specified"))
	}

	if c.Endpoint == "" {
		c.Endpoint = defaultEndpoint
	}

	if c.Zone == "" {
		c.Zone = defaultZone
	}

	// provision config by OS environment variables
	if c.Token == "" {
		c.Token = os.Getenv("YC_TOKEN")
	}

	if c.ServiceAccountKeyFile == "" {
		c.ServiceAccountKeyFile = os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")
	}

	if c.FolderID == "" {
		c.FolderID = os.Getenv("YC_FOLDER_ID")
	}

	if c.Token == "" && c.ServiceAccountKeyFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a token or service account key file must be specified"))
	}

	if c.Token != "" && c.ServiceAccountKeyFile != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("one of token or service account key file must be specified, not both"))
	}

	if c.Token != "" {
		packer.LogSecretFilter.Set(c.Token)
	}

	if c.ServiceAccountKeyFile != "" {
		if _, err := iamkey.ReadFromJSONFile(c.ServiceAccountKeyFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("fail to read service account key file: %s", err))
		}
	}

	if c.FolderID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a folder_id must be specified"))
	}

	if c.StateTimeout == 0 {
		c.StateTimeout = 5 * time.Minute
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

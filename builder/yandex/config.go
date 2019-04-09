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

	Endpoint              string `mapstructure:"endpoint"`
	FolderID              string `mapstructure:"folder_id"`
	ServiceAccountKeyFile string `mapstructure:"service_account_key_file"`
	Token                 string `mapstructure:"token"`

	DiskName            string            `mapstructure:"disk_name"`
	DiskSizeGb          int               `mapstructure:"disk_size_gb"`
	DiskType            string            `mapstructure:"disk_type"`
	ImageDescription    string            `mapstructure:"image_description"`
	ImageFamily         string            `mapstructure:"image_family"`
	ImageLabels         map[string]string `mapstructure:"image_labels"`
	ImageName           string            `mapstructure:"image_name"`
	ImageProductIDs     []string          `mapstructure:"image_product_ids"`
	InstanceCores       int               `mapstructure:"instance_cores"`
	InstanceMemory      int               `mapstructure:"instance_mem_gb"`
	InstanceName        string            `mapstructure:"instance_name"`
	Labels              map[string]string `mapstructure:"labels"`
	PlatformID          string            `mapstructure:"platform_id"`
	Metadata            map[string]string `mapstructure:"metadata"`
	SerialLogFile       string            `mapstructure:"serial_log_file"`
	SourceImageFamily   string            `mapstructure:"source_image_family"`
	SourceImageFolderID string            `mapstructure:"source_image_folder_id"`
	SourceImageID       string            `mapstructure:"source_image_id"`
	SubnetID            string            `mapstructure:"subnet_id"`
	UseIPv4Nat          bool              `mapstructure:"use_ipv4_nat"`
	UseIPv6             bool              `mapstructure:"use_ipv6"`
	UseInternalIP       bool              `mapstructure:"use_internal_ip"`
	Zone                string            `mapstructure:"zone"`

	ctx          interpolate.Context
	StateTimeout time.Duration `mapstructure:"state_timeout"`
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

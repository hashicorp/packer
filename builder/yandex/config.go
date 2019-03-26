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

var reImageFamily = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Communicator        communicator.Config `mapstructure:",squash"`

	Endpoint              string `mapstructure:"endpoint"`
	Token                 string `mapstructure:"token"`
	ServiceAccountKeyFile string `mapstructure:"service_account_key_file"`
	FolderID              string `mapstructure:"folder_id"`
	Zone                  string `mapstructure:"zone"`

	SerialLogFile       string            `mapstructure:"serial_log_file"`
	InstanceCores       int               `mapstructure:"instance_cores"`
	InstanceMemory      int               `mapstructure:"instance_mem_gb"`
	DiskSizeGb          int               `mapstructure:"disk_size_gb"`
	DiskType            string            `mapstructure:"disk_type"`
	SubnetID            string            `mapstructure:"subnet_id"`
	ImageName           string            `mapstructure:"image_name"`
	ImageFamily         string            `mapstructure:"image_family"`
	ImageDescription    string            `mapstructure:"image_description"`
	ImageLabels         map[string]string `mapstructure:"image_labels"`
	ImageProductIDs     []string          `mapstructure:"image_product_ids"`
	InstanceName        string            `mapstructure:"instance_name"`
	Labels              map[string]string `mapstructure:"labels"`
	DiskName            string            `mapstructure:"disk_name"`
	MachineType         string            `mapstructure:"machine_type"`
	Metadata            map[string]string `mapstructure:"metadata"`
	SourceImageID       string            `mapstructure:"source_image_id"`
	SourceImageFamily   string            `mapstructure:"source_image_family"`
	SourceImageFolderID string            `mapstructure:"source_image_folder_id"`
	UseInternalIP       bool              `mapstructure:"use_internal_ip"`
	UseIPv4Nat          bool              `mapstructure:"use_ipv4_nat"`
	UseIPv6             bool              `mapstructure:"use_ipv6"`

	RawStepTimeout string `mapstructure:"step_timeout"`

	stepTimeout time.Duration
	ctx         interpolate.Context
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

	if c.MachineType == "" {
		c.MachineType = "standard-v1"
	}

	if c.RawStepTimeout == "" {
		c.RawStepTimeout = "5m"
	}

	if es := c.Communicator.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.

	if c.SourceImageID == "" && c.SourceImageFamily == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image_id or source_image_family must be specified"))
	}

	err = c.CalcTimeout()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if c.Endpoint == "" {
		c.Endpoint = "api.cloud.yandex.net:443"
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
				errs, fmt.Errorf("fail to parse service account key file: %s", err))
		}

	}

	if c.FolderID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a folder_id must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

func (c *Config) CalcTimeout() error {
	stepTimeout, err := time.ParseDuration(c.RawStepTimeout)
	if err != nil {
		return fmt.Errorf("Failed parsing step_timeout: %s", err)
	}
	c.stepTimeout = stepTimeout
	return nil
}

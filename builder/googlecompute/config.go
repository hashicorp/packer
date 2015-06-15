package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the configuration structure for the GCE builder. It stores
// both the publicly settable state as well as the privately generated
// state of the config object.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	AccountFile string `mapstructure:"account_file"`
	ProjectId   string `mapstructure:"project_id"`

	DiskName             string            `mapstructure:"disk_name"`
	DiskSizeGb           int64             `mapstructure:"disk_size"`
	ImageName            string            `mapstructure:"image_name"`
	ImageDescription     string            `mapstructure:"image_description"`
	InstanceName         string            `mapstructure:"instance_name"`
	MachineType          string            `mapstructure:"machine_type"`
	Metadata             map[string]string `mapstructure:"metadata"`
	Network              string            `mapstructure:"network"`
	SourceImage          string            `mapstructure:"source_image"`
	SourceImageProjectId string            `mapstructure:"source_image_project_id"`
	RawStateTimeout      string            `mapstructure:"state_timeout"`
	Tags                 []string          `mapstructure:"tags"`
	Zone                 string            `mapstructure:"zone"`

	account         accountFile
	privateKeyBytes []byte
	stateTimeout    time.Duration
	ctx             *interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Set defaults.
	if c.Network == "" {
		c.Network = "default"
	}

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 10
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		c.ImageName = "packer-{{timestamp}}"
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName
	}

	if c.MachineType == "" {
		c.MachineType = "n1-standard-1"
	}

	if c.RawStateTimeout == "" {
		c.RawStateTimeout = "5m"
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	var errs *packer.MultiError

	// Process required parameters.
	if c.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}

	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}

	stateTimeout, err := time.ParseDuration(c.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	c.stateTimeout = stateTimeout

	if c.AccountFile != "" {
		if err := loadJSON(&c.account, c.AccountFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed parsing account file: %s", err))
		}
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

package profitbricks

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"os"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	PBUsername string `mapstructure:"username"`
	PBPassword string `mapstructure:"password"`
	PBUrl      string `mapstructure:"url"`

	Region           string `mapstructure:"location"`
	Image            string `mapstructure:"image"`
	SSHKey           string
	SSHKey_path      string              `mapstructure:"ssh_key_path"`
	SnapshotName     string              `mapstructure:"snapshot_name"`
	SnapshotPassword string              `mapstructure:"snapshot_password"`
	DiskSize         int                 `mapstructure:"disk_size"`
	DiskType         string              `mapstructure:"disk_type"`
	Cores            int                 `mapstructure:"cores"`
	Ram              int                 `mapstructure:"ram"`
	Timeout          int                 `mapstructure:"timeout"`
	CommConfig       communicator.Config `mapstructure:",squash"`
	ctx              interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}
	c.Comm.SSHPort = 22

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.PBUsername == "" {
		c.PBUsername = os.Getenv("PROFITBRICKS_USERNAME")
	}

	if c.PBPassword == "" {
		c.PBPassword = os.Getenv("PROFITBRICKS_PASSWORD")
	}

	if c.PBUrl == "" {
		c.PBUrl = "https://api.profitbricks.com/rest/v2"
	}

	if c.Image == "" {
		c.Image = "Ubuntu-16.04"
	}

	if c.Cores == 0 {
		c.Cores = 4
	}

	if c.Ram == 0 {
		c.Ram = 2048
	}

	if c.DiskSize == 0 {
		c.DiskSize = 50
	}

	if c.Region == "" {
		c.Region = "us/las"
	}

	if c.DiskType == "" {
		c.DiskType = "HDD"
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	c.Comm.SSHPort = 22

	if c.PBUsername == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ProfitBricks username is required"))
	}

	if c.PBPassword == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ProfitBricks password is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}
	common.ScrubConfig(c, c.PBUsername)

	return &c, nil, nil
}

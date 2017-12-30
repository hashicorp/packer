package profitbricks

import (
	"errors"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
	"os"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	PBUsername string `mapstructure:"username"`
	PBPassword string `mapstructure:"password"`
	PBUrl      string `mapstructure:"url"`

	Region       string `mapstructure:"location"`
	Image        string `mapstructure:"image"`
	SSHKey       string
	SnapshotName string              `mapstructure:"snapshot_name"`
	DiskSize     int                 `mapstructure:"disk_size"`
	DiskType     string              `mapstructure:"disk_type"`
	Cores        int                 `mapstructure:"cores"`
	Ram          int                 `mapstructure:"ram"`
	Retries      int                 `mapstructure:"retries"`
	CommConfig   communicator.Config `mapstructure:",squash"`
	ctx          interpolate.Context
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

	if c.Comm.SSHPassword == "" && c.Comm.SSHPrivateKey == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Either ssh private key path or ssh password must be set."))
	}

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
		c.PBUrl = "https://api.profitbricks.com/cloudapi/v4"
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

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ProfitBricks 'image' is required"))
	}

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

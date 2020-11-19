//go:generate mapstructure-to-hcl2 -type Config

package profitbricks

import (
	"errors"
	"os"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
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
	SnapshotName string `mapstructure:"snapshot_name"`
	DiskSize     int    `mapstructure:"disk_size"`
	DiskType     string `mapstructure:"disk_type"`
	Cores        int    `mapstructure:"cores"`
	Ram          int    `mapstructure:"ram"`
	Retries      int    `mapstructure:"retries"`
	ctx          interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

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
		return nil, err
	}

	var errs *packersdk.MultiError

	if c.Comm.SSHPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
		errs = packersdk.MultiErrorAppend(
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
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("ProfitBricks 'image' is required"))
	}

	if c.PBUsername == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("ProfitBricks username is required"))
	}

	if c.PBPassword == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("ProfitBricks password is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}
	packer.LogSecretFilter.Set(c.PBUsername)

	return nil, nil
}

package oneandone

import (
	"github.com/1and1/oneandone-cloudserver-sdk-go"
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

	Token         string `mapstructure:"token"`
	Url           string `mapstructure:"url"`
	SSHKey        string
	SSHKey_path   string              `mapstructure:"ssh_key_path"`
	SnapshotName  string              `mapstructure:"image_name"`
	Image         string              `mapstructure:"source_image_name"`
	ImagePassword string              `mapstructure:"image_password"`
	DiskSize      int                 `mapstructure:"disk_size"`
	Timeout       int                 `mapstructure:"timeout"`
	CommConfig    communicator.Config `mapstructure:",squash"`
	ctx           interpolate.Context
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

	if c.Token == "" {
		c.Token = os.Getenv("ONEANDONE_TOKEN")
	}

	if c.Url == "" {
		c.Url = oneandone.BaseUrl
	}

	if c.DiskSize == 0 {
		c.DiskSize = 50
	}

	if c.Image == "" {
		c.Image = "ubuntu1604-64std"
	}

	if c.Timeout == 0 {
		c.Timeout = 600
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}
	common.ScrubConfig(c, c.Token)

	return &c, nil, nil
}

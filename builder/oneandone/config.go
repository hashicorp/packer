package oneandone

import (
	"errors"
	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"os"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Token          string `mapstructure:"token"`
	Url            string `mapstructure:"url"`
	SSHKey         string
	SnapshotName   string `mapstructure:"image_name"`
	DataCenterName string `mapstructure:"data_center_name"`
	DataCenterId   string
	Image          string              `mapstructure:"source_image_name"`
	DiskSize       int                 `mapstructure:"disk_size"`
	Retries        int                 `mapstructure:"retries"`
	CommConfig     communicator.Config `mapstructure:",squash"`
	ctx            interpolate.Context
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

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("1&1 'image' is required"))
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	if c.Token == "" {
		c.Token = os.Getenv("ONEANDONE_TOKEN")
	}

	if c.Url == "" {
		c.Url = oneandone.BaseUrl
	}

	if c.DiskSize == 0 {
		c.DiskSize = 50
	}

	if c.Retries == 0 {
		c.Retries = 600
	}

	if c.DataCenterName != "" {
		token := oneandone.SetToken(c.Token)

		//Create an API client
		api := oneandone.New(token, c.Url)

		dcs, err := api.ListDatacenters()

		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, err)
		}
		for _, dc := range dcs {
			if strings.ToLower(dc.CountryCode) == strings.ToLower(c.DataCenterName) {
				c.DataCenterId = dc.Id
				break
			}
		}
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

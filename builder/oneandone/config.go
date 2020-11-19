//go:generate mapstructure-to-hcl2 -type Config

package oneandone

import (
	"errors"
	"os"
	"strings"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
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

	Token          string `mapstructure:"token"`
	Url            string `mapstructure:"url"`
	SnapshotName   string `mapstructure:"image_name"`
	DataCenterName string `mapstructure:"data_center_name"`
	DataCenterId   string
	Image          string `mapstructure:"source_image_name"`
	DiskSize       int    `mapstructure:"disk_size"`
	Retries        int    `mapstructure:"retries"`
	ctx            interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
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

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("1&1 'image' is required"))
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
			errs = packersdk.MultiErrorAppend(
				errs, err)
		}
		for _, dc := range dcs {
			if strings.EqualFold(dc.CountryCode, c.DataCenterName) {
				c.DataCenterId = dc.Id
				break
			}
		}
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}
	packer.LogSecretFilter.Set(c.Token)
	return nil, nil
}

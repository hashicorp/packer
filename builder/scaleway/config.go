package scaleway

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Token        string `mapstructure:"api_token"`
	Organization string `mapstructure:"api_access_key"`

	Region         string `mapstructure:"region"`
	Image          string `mapstructure:"image"`
	CommercialType string `mapstructure:"commercial_type"`

	SnapshotName string `mapstructure:"snapshot_name"`
	ImageName    string `mapstructure:"image_name"`
	ServerName   string `mapstructure:"server_name"`

	UserAgent string
	ctx       interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

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
		return nil, nil, err
	}

	c.UserAgent = "Packer - Scaleway builder"

	if c.Organization == "" {
		c.Organization = os.Getenv("SCALEWAY_API_ACCESS_KEY")
	}

	if c.Token == "" {
		c.Token = os.Getenv("SCALEWAY_API_TOKEN")
	}

	if c.SnapshotName == "" {
		def, err := interpolate.Render("snapshot-packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		c.SnapshotName = def
	}

	if c.ImageName == "" {
		def, err := interpolate.Render("image-packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		c.ImageName = def
	}

	if c.ServerName == "" {
		// Default to packer-[time-ordered-uuid]
		c.ServerName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.Organization == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Scaleway Organization ID must be specified"))
	}

	if c.Token == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Scaleway Token must be specified"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.CommercialType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("commercial type is required"))
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	common.ScrubConfig(c, c.Token)
	return c, nil, nil
}

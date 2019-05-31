//go:generate struct-markdown

package scaleway

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/useragent"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	// The token to use to authenticate with your account.
    // It can also be specified via environment variable SCALEWAY_API_TOKEN. You
    // can see and generate tokens in the "Credentials"
    // section of the control panel.
	Token        string `mapstructure:"api_token" required:"true"`
	// The organization id to use to identify your
    // organization. It can also be specified via environment variable
    // SCALEWAY_ORGANIZATION. Your organization id is available in the
    // "Account" section of the
    // control panel.
    // Previously named: api_access_key with environment variable: SCALEWAY_API_ACCESS_KEY
	Organization string `mapstructure:"organization_id" required:"true"`
	// The name of the region to launch the server in (par1
    // or ams1). Consequently, this is the region where the snapshot will be
    // available.
	Region         string `mapstructure:"region" required:"true"`
	// The UUID of the base image to use. This is the image
    // that will be used to launch a new server and provision it. See
    // the images list
    // get the complete list of the accepted image UUID.
	Image          string `mapstructure:"image" required:"true"`
	// The name of the server commercial type:
    // ARM64-128GB, ARM64-16GB, ARM64-2GB, ARM64-32GB, ARM64-4GB,
    // ARM64-64GB, ARM64-8GB, C1, C2L, C2M, C2S, START1-L,
    // START1-M, START1-S, START1-XS, X64-120GB, X64-15GB, X64-30GB,
    // X64-60GB
	CommercialType string `mapstructure:"commercial_type" required:"true"`
	// The name of the resulting snapshot that will
    // appear in your account. Default packer-TIMESTAMP
	SnapshotName string `mapstructure:"snapshot_name" required:"false"`
	// The name of the resulting image that will appear in
    // your account. Default packer-TIMESTAMP
	ImageName    string `mapstructure:"image_name" required:"false"`
	// The name assigned to the server. Default
    // packer-UUID
	ServerName   string `mapstructure:"server_name" required:"false"`
	// The id of an existing bootscript to use when
    // booting the server.
	Bootscript   string `mapstructure:"bootscript" required:"false"`
	// The type of boot, can be either local or
    // bootscript, Default bootscript
	BootType     string `mapstructure:"boottype" required:"false"`

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

	c.UserAgent = useragent.String()

	if c.Organization == "" {
		if os.Getenv("SCALEWAY_ORGANIZATION") != "" {
			c.Organization = os.Getenv("SCALEWAY_ORGANIZATION")
		} else {
			log.Printf("Deprecation warning: Use SCALEWAY_ORGANIZATION environment variable and organization_id argument instead of api_access_key argument and SCALEWAY_API_ACCESS_KEY environment variable.")
			c.Organization = os.Getenv("SCALEWAY_API_ACCESS_KEY")
		}
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

	if c.BootType == "" {
		c.BootType = "bootscript"
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

	packer.LogSecretFilter.Set(c.Token)
	return c, nil, nil
}

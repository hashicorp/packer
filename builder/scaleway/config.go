//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package scaleway

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/hashicorp/packer/builder/scaleway/version"
	"github.com/mitchellh/mapstructure"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	// The AccessKey corresponding to the secret key.
	// Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
	// It can also be specified via the environment variable SCW_ACCESS_KEY.
	AccessKey string `mapstructure:"access_key" required:"true"`
	// The SecretKey to authenticate against the Scaleway API.
	// Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
	// It can also be specified via the environment variable SCW_SECRET_KEY.
	SecretKey string `mapstructure:"secret_key" required:"true"`
	// The Project ID in which the instances, volumes and snapshots will be created.
	// Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
	// It can also be specified via the environment variable SCW_DEFAULT_PROJECT_ID.
	ProjectID string `mapstructure:"project_id" required:"true"`
	// The Zone in which the instances, volumes and snapshots will be created.
	// Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
	// It can also be specified via the environment variable SCW_DEFAULT_ZONE
	Zone string `mapstructure:"zone" required:"true"`
	// The Scaleway API URL to use
	// Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
	// It can also be specified via the environment variable SCW_API_URL
	APIURL string `mapstructure:"api_url"`

	// The UUID of the base image to use. This is the image
	// that will be used to launch a new server and provision it. See
	// the images list
	// get the complete list of the accepted image UUID.
	// The marketplace image label (eg `ubuntu_focal`) also works.
	Image string `mapstructure:"image" required:"true"`
	// The name of the server commercial type:
	// C1, C2L, C2M, C2S, DEV1-S, DEV1-M, DEV1-L, DEV1-XL,
	// GP1-XS, GP1-S, GP1-M, GP1-L, GP1-XL, RENDER-S
	CommercialType string `mapstructure:"commercial_type" required:"true"`
	// The name of the resulting snapshot that will
	// appear in your account. Default packer-TIMESTAMP
	SnapshotName string `mapstructure:"snapshot_name" required:"false"`
	// The name of the resulting image that will appear in
	// your account. Default packer-TIMESTAMP
	ImageName string `mapstructure:"image_name" required:"false"`
	// The name assigned to the server. Default
	// packer-UUID
	ServerName string `mapstructure:"server_name" required:"false"`
	// The id of an existing bootscript to use when
	// booting the server.
	Bootscript string `mapstructure:"bootscript" required:"false"`
	// The type of boot, can be either local or
	// bootscript, Default bootscript
	BootType string `mapstructure:"boottype" required:"false"`

	RemoveVolume bool `mapstructure:"remove_volume"`

	// Shutdown timeout. Default to 5m
	ShutdownTimeout string `mapstructure:"shutdown_timeout" required:"false"`
	UserAgent       string `mapstructure-to-hcl2:",skip"`
	ctx             interpolate.Context

	// Deprecated configs

	// The token to use to authenticate with your account.
	// It can also be specified via environment variable SCALEWAY_API_TOKEN. You
	// can see and generate tokens in the "Credentials"
	// section of the control panel.
	// Deprecated, use SecretKey instead
	Token string `mapstructure:"api_token" required:"false"`
	// The organization id to use to identify your
	// organization. It can also be specified via environment variable
	// SCALEWAY_ORGANIZATION. Your organization id is available in the
	// "Account" section of the
	// control panel.
	// Previously named: api_access_key with environment variable: SCALEWAY_API_ACCESS_KEY
	// Deprecated, use ProjectID instead
	Organization string `mapstructure:"organization_id" required:"false"`
	// The name of the region to launch the server in (par1
	// or ams1). Consequently, this is the region where the snapshot will be
	// available.
	// Deprecated, use Zone instead
	Region string `mapstructure:"region" required:"false"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		PluginType:         BuilderId,
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

	var warnings []string

	c.UserAgent = useragent.String(version.ScalewayPluginVersion.FormattedVersion())

	configFile, err := scw.LoadConfig()
	// If the config file do not exist, don't return an error as we may find config in ENV or flags.
	if _, isNotFoundError := err.(*scw.ConfigFileNotFoundError); isNotFoundError {
		configFile = &scw.Config{}
	} else if err != nil {
		return nil, err
	}
	activeProfile, err := configFile.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	envProfile := scw.LoadEnvProfile()
	profile := scw.MergeProfiles(activeProfile, envProfile)

	// Deprecated variables
	if c.Organization == "" {
		if os.Getenv("SCALEWAY_ORGANIZATION") != "" {
			c.Organization = os.Getenv("SCALEWAY_ORGANIZATION")
		} else {
			log.Printf("Deprecation warning: Use SCALEWAY_ORGANIZATION environment variable and organization_id argument instead of api_access_key argument and SCALEWAY_API_ACCESS_KEY environment variable.")
			c.Organization = os.Getenv("SCALEWAY_API_ACCESS_KEY")
		}
	}
	if c.Organization != "" {
		warnings = append(warnings, "organization_id is deprecated in favor of project_id")
		c.ProjectID = c.Organization
	}

	if c.Token == "" {
		c.Token = os.Getenv("SCALEWAY_API_TOKEN")
	}
	if c.Token != "" {
		warnings = append(warnings, "token is deprecated in favor of secret_key")
		c.SecretKey = c.Token
	}

	if c.Region != "" {
		warnings = append(warnings, "region is deprecated in favor of zone")
		c.Zone = c.Region
	}

	if c.AccessKey == "" {
		if profile.AccessKey != nil {
			c.AccessKey = *profile.AccessKey
		}
	}

	if c.SecretKey == "" {
		if profile.SecretKey != nil {
			c.SecretKey = *profile.SecretKey
		}
	}

	if c.ProjectID == "" {
		if profile.DefaultProjectID != nil {
			c.ProjectID = *profile.DefaultProjectID
		}
	}

	if c.Zone == "" {
		if profile.DefaultZone != nil {
			c.Zone = *profile.DefaultZone
		}
	}

	if c.APIURL == "" {
		if profile.APIURL != nil {
			c.APIURL = *profile.APIURL
		}
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
		c.BootType = instance.BootTypeLocal.String()
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	if c.ProjectID == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("scaleway Project ID must be specified"))
	}

	if c.SecretKey == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("scaleway Secret Key must be specified"))
	}

	if c.AccessKey == "" {
		warnings = append(warnings, "access_key will be required in future versions")
		c.AccessKey = "SCWXXXXXXXXXXXXXXXXX"
	}

	if c.Zone == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("Scaleway Zone is required"))
	}

	if c.CommercialType == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("commercial type is required"))
	}

	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if c.ShutdownTimeout == "" {
		c.ShutdownTimeout = "5m"
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	packersdk.LogSecretFilter.Set(c.Token)
	packersdk.LogSecretFilter.Set(c.SecretKey)
	return warnings, nil
}

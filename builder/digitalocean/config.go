package digitalocean

import (
	"errors"
	"fmt"
	"os"
	"time"

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

	APIToken string `mapstructure:"api_token"`
	APIURL   string `mapstructure:"api_url"`

	Region string `mapstructure:"region"`
	Size   string `mapstructure:"size"`
	Image  string `mapstructure:"image"`

	PrivateNetworking bool           `mapstructure:"private_networking"`
	Monitoring        bool           `mapstructure:"monitoring"`
	SnapshotName      string         `mapstructure:"snapshot_name"`
	SnapshotRegions   []string       `mapstructure:"snapshot_regions"`
	StateTimeout      time.Duration  `mapstructure:"state_timeout"`
	DropletName       string         `mapstructure:"droplet_name"`
	UserData          string         `mapstructure:"user_data"`
	UserDataFile      string         `mapstructure:"user_data_file"`
	Volumes           []VolumeConfig `mapstructure:"volumes"`

	ctx interpolate.Context
}

type VolumeConfig struct {
	VolumeName     string `mapstructure:"volume_name"`
	Size           int64  `mapstructure:"size"`
	BaseSnapshotID string `mapstructure:"base_snapshot_id"`
	SnapshotName   string `mapstructure:"snapshot_name"`
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

	// Defaults
	if c.APIToken == "" {
		// Default to environment variable for api_token, if it exists
		c.APIToken = os.Getenv("DIGITALOCEAN_API_TOKEN")
	}
	if c.APIURL == "" {
		c.APIURL = os.Getenv("DIGITALOCEAN_API_URL")
	}
	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.DropletName == "" {
		// Default to packer-[time-ordered-uuid]
		c.DropletName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.StateTimeout == 0 {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		c.StateTimeout = 6 * time.Minute
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.APIToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("api_token for auth must be specified"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.Size == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("size is required"))
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, errors.New(fmt.Sprintf("user_data_file not found: %s", c.UserDataFile)))
		}
	}

	for i := 0; i < len(c.Volumes); i++ {
		if c.Volumes[i].VolumeName == "" {
			c.Volumes[i].VolumeName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
		}

		if c.Volumes[i].SnapshotName == "" {
			c.Volumes[i].SnapshotName = fmt.Sprintf("%s-vol%d", c.SnapshotName, i)
		}

		if c.Volumes[i].Size == 0 {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("volume %d: size is required", i))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	common.ScrubConfig(c, c.APIToken)
	return c, nil, nil
}

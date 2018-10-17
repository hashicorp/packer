package hcloud

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	HCloudToken  string        `mapstructure:"token"`
	Endpoint     string        `mapstructure:"endpoint"`
	PollInterval time.Duration `mapstructure:"poll_interval"`

	ServerName string `mapstructure:"server_name"`
	Location   string `mapstructure:"location"`
	ServerType string `mapstructure:"server_type"`
	Image      string `mapstructure:"image"`

	SnapshotName string `mapstructure:"snapshot_name"`
	UserData     string `mapstructure:"user_data"`
	UserDataFile string `mapstructure:"user_data_file"`

	ctx interpolate.Context
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
	if c.HCloudToken == "" {
		c.HCloudToken = os.Getenv("HCLOUD_TOKEN")
	}
	if c.Endpoint == "" {
		if os.Getenv("HCLOUD_ENDPOINT") != "" {
			c.Endpoint = os.Getenv("HCLOUD_ENDPOINT")
		} else {
			c.Endpoint = hcloud.Endpoint
		}
	}
	if c.PollInterval == 0 {
		c.PollInterval = 500 * time.Millisecond
	}

	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}
		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.ServerName == "" {
		// Default to packer-[time-ordered-uuid]
		c.ServerName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.HCloudToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("token for auth must be specified"))
	}

	if c.Location == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("location is required"))
	}

	if c.ServerType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("server type is required"))
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

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.HCloudToken)
	return c, nil, nil
}

func getServerIP(state multistep.StateBag) (string, error) {
	return state.Get("server_ip").(string), nil
}

package digitalocean

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	//"github.com/digitalocean/godo"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	APIToken string `mapstructure:"api_token"`

	// OLD STUFF

	Region string `mapstructure:"region"`
	Size   string `mapstructure:"size"`
	Image  string `mapstructure:"image"`

	PrivateNetworking bool   `mapstructure:"private_networking"`
	SnapshotName      string `mapstructure:"snapshot_name"`
	DropletName       string `mapstructure:"droplet_name"`
	SSHUsername       string `mapstructure:"ssh_username"`
	SSHPort           uint   `mapstructure:"ssh_port"`

	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawStateTimeout string `mapstructure:"state_timeout"`

	// These are unexported since they're set by other fields
	// being set.
	sshTimeout   time.Duration
	stateTimeout time.Duration

	ctx *interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		Metadata:    &md,
		Interpolate: true,
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

	if c.Region == "" {
		c.Region = DefaultRegion
	}

	if c.Size == "" {
		c.Size = DefaultSize
	}

	if c.Image == "" {
		c.Image = DefaultImage
	}

	if c.SnapshotName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = "packer-{{timestamp}}"
	}

	if c.DropletName == "" {
		// Default to packer-[time-ordered-uuid]
		c.DropletName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.SSHUsername == "" {
		// Default to "root". You can override this if your
		// SourceImage has a different user account then the DO default
		c.SSHUsername = "root"
	}

	if c.SSHPort == 0 {
		// Default to port 22 per DO default
		c.SSHPort = 22
	}

	if c.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		c.RawSSHTimeout = "1m"
	}

	if c.RawStateTimeout == "" {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		c.RawStateTimeout = "6m"
	}

	var errs *packer.MultiError
	if c.APIToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("api_token for auth must be specified"))
	}

	sshTimeout, err := time.ParseDuration(c.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	c.sshTimeout = sshTimeout

	stateTimeout, err := time.ParseDuration(c.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	c.stateTimeout = stateTimeout

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	common.ScrubConfig(c, c.APIToken)
	return &c, nil, nil
}

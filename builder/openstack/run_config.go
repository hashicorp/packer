package openstack

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	SourceImage      string   `mapstructure:"source_image"`
	Flavor           string   `mapstructure:"flavor"`
	RawSSHTimeout    string   `mapstructure:"ssh_timeout"`
	SSHUsername      string   `mapstructure:"ssh_username"`
	SSHPort          int      `mapstructure:"ssh_port"`
	SSHInterface     string   `mapstructure:"ssh_interface"`
	AvailabilityZone string   `mapstructure:"availability_zone"`
	RackconnectWait  bool     `mapstructure:"rackconnect_wait"`
	FloatingIpPool   string   `mapstructure:"floating_ip_pool"`
	FloatingIp       string   `mapstructure:"floating_ip"`
	SecurityGroups   []string `mapstructure:"security_groups"`
	Networks         []string `mapstructure:"networks"`
	UserData         string   `mapstructure:"user_data"`
	UserDataFile     string   `mapstructure:"user_data_file"`

	// Not really used, but here for BC
	OpenstackProvider string `mapstructure:"openstack_provider"`
	UseFloatingIp     bool   `mapstructure:"use_floating_ip"`

	// Unexported fields that are calculated from others
	sshTimeout time.Duration
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// Defaults
	if c.SSHUsername == "" {
		c.SSHUsername = "root"
	}

	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHTimeout == "" {
		c.RawSSHTimeout = "5m"
	}

	if c.UseFloatingIp && c.FloatingIpPool == "" {
		c.FloatingIpPool = "public"
	}

	// Validation
	var err error
	errs := make([]error, 0)
	if c.SourceImage == "" {
		errs = append(errs, errors.New("A source_image must be specified"))
	}

	if c.Flavor == "" {
		errs = append(errs, errors.New("A flavor must be specified"))
	}

	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	c.sshTimeout, err = time.ParseDuration(c.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	return errs
}

func (c *RunConfig) SSHTimeout() time.Duration {
	return c.sshTimeout
}

package openstack

import (
	"errors"

	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	Comm         communicator.Config `mapstructure:",squash"`
	SSHInterface string              `mapstructure:"ssh_interface"`

	SourceImage      string   `mapstructure:"source_image"`
	Flavor           string   `mapstructure:"flavor"`
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
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// Defaults
	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	if c.UseFloatingIp && c.FloatingIpPool == "" {
		c.FloatingIpPool = "public"
	}

	// Validation
	errs := c.Comm.Prepare(ctx)
	if c.SourceImage == "" {
		errs = append(errs, errors.New("A source_image must be specified"))
	}

	if c.Flavor == "" {
		errs = append(errs, errors.New("A flavor must be specified"))
	}

	return errs
}

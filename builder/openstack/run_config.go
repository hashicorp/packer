package openstack

import (
	"errors"

	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	Comm           communicator.Config `mapstructure:",squash"`
	SSHKeyPairName string              `mapstructure:"ssh_keypair_name"`
	SSHInterface   string              `mapstructure:"ssh_interface"`
	SSHIPVersion   string              `mapstructure:"ssh_ip_version"`

	SourceImage      string   `mapstructure:"source_image"`
	SourceImageName  string   `mapstructure:"source_image_name"`
	Flavor           string   `mapstructure:"flavor"`
	AvailabilityZone string   `mapstructure:"availability_zone"`
	RackconnectWait  bool     `mapstructure:"rackconnect_wait"`
	FloatingIpPool   string   `mapstructure:"floating_ip_pool"`
	FloatingIp       string   `mapstructure:"floating_ip"`
	SecurityGroups   []string `mapstructure:"security_groups"`
	Networks         []string `mapstructure:"networks"`
	UserData         string   `mapstructure:"user_data"`
	UserDataFile     string   `mapstructure:"user_data_file"`

	ConfigDrive bool `mapstructure:"config_drive"`

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
	if c.SourceImage == "" && c.SourceImageName == "" {
		errs = append(errs, errors.New("Either a source_image or a source_image_name must be specified"))
	} else if len(c.SourceImage) > 0 && len(c.SourceImageName) > 0 {
		errs = append(errs, errors.New("Only a source_image or a source_image_name can be specified, not both."))
	}

	if c.Flavor == "" {
		errs = append(errs, errors.New("A flavor must be specified"))
	}

	if c.SSHIPVersion != "" && c.SSHIPVersion != "4" && c.SSHIPVersion != "6" {
		errs = append(errs, errors.New("SSH IP version must be either 4 or 6"))
	}

	return errs
}

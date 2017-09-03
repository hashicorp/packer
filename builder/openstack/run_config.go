package openstack

import (
	"errors"
	"fmt"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	Comm                 communicator.Config `mapstructure:",squash"`
	SSHKeyPairName       string              `mapstructure:"ssh_keypair_name"`
	TemporaryKeyPairName string              `mapstructure:"temporary_key_pair_name"`
	SSHInterface         string              `mapstructure:"ssh_interface"`
	SSHIPVersion         string              `mapstructure:"ssh_ip_version"`

	SourceImage      string            `mapstructure:"source_image"`
	SourceImageName  string            `mapstructure:"source_image_name"`
	Flavor           string            `mapstructure:"flavor"`
	AvailabilityZone string            `mapstructure:"availability_zone"`
	RackconnectWait  bool              `mapstructure:"rackconnect_wait"`
	FloatingIpPool   string            `mapstructure:"floating_ip_pool"`
	FloatingIp       string            `mapstructure:"floating_ip"`
	ReuseIps         bool              `mapstructure:"reuse_ips"`
	SecurityGroups   []string          `mapstructure:"security_groups"`
	Networks         []string          `mapstructure:"networks"`
	UserData         string            `mapstructure:"user_data"`
	UserDataFile     string            `mapstructure:"user_data_file"`
	InstanceMetadata map[string]string `mapstructure:"instance_metadata"`

	ConfigDrive bool `mapstructure:"config_drive"`

	// Not really used, but here for BC
	OpenstackProvider string `mapstructure:"openstack_provider"`
	UseFloatingIp     bool   `mapstructure:"use_floating_ip"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided and we are not using
	// ssh_password.
	if c.SSHKeyPairName == "" && c.TemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKey == "" && c.Comm.SSHPassword == "" {

		c.TemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.UseFloatingIp && c.FloatingIpPool == "" {
		c.FloatingIpPool = "public"
	}

	// Validation
	errs := c.Comm.Prepare(ctx)

	if c.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKey == "" {
			errs = append(errs, errors.New("A ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKey == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, errors.New("A ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

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

	for key, value := range c.InstanceMetadata {
		if len(key) > 255 {
			errs = append(errs, fmt.Errorf("Instance metadata key too long (max 255 bytes): %s", key))
		}
		if len(value) > 255 {
			errs = append(errs, fmt.Errorf("Instance metadata value too long (max 255 bytes): %s", value))
		}
	}

	return errs
}

package openstack

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"time"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	SourceImage       string   `mapstructure:"source_image"`
	Flavor            string   `mapstructure:"flavor"`
	RawSSHTimeout     string   `mapstructure:"ssh_timeout"`
	SSHUsername       string   `mapstructure:"ssh_username"`
	SSHPort           int      `mapstructure:"ssh_port"`
	OpenstackProvider string   `mapstructure:"openstack_provider"`
	UseFloatingIp     bool     `mapstructure:"use_floating_ip"`
	FloatingIpPool    string   `mapstructure:"floating_ip_pool"`
	FloatingIp        string   `mapstructure:"floating_ip"`
	SecurityGroups    []string `mapstructure:"security_groups"`

	// Unexported fields that are calculated from others
	sshTimeout time.Duration
}

func (c *RunConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

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

	templates := map[string]*string{
		"flavlor":      &c.Flavor,
		"ssh_timeout":  &c.RawSSHTimeout,
		"ssh_username": &c.SSHUsername,
		"source_image": &c.SourceImage,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
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

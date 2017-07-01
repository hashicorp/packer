package main

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"time"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Connection
	VCenterServer      string `mapstructure:"vcenter_server"`
	Datacenter         string `mapstructure:"datacenter"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	InsecureConnection bool   `mapstructure:"insecure_connection"`

	// Location
	Template     string `mapstructure:"template"`
	FolderName   string `mapstructure:"folder"`
	VMName       string `mapstructure:"vm_name"`
	Host         string `mapstructure:"host"`
	ResourcePool string `mapstructure:"resource_pool"`
	Datastore    string `mapstructure:"datastore"`
	LinkedClone  bool   `mapstructure:"linked_clone"`

	// Customization
	HardwareConfig `mapstructure:",squash"`

	// Provisioning
	communicator.Config `mapstructure:",squash"`

	// Post-processing
	ShutdownCommand    string `mapstructure:"shutdown_command"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	ShutdownTimeout    time.Duration
	CreateSnapshot     bool   `mapstructure:"create_snapshot"`
	ConvertToTemplate  bool   `mapstructure:"convert_to_template"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	{
		err := config.Decode(c, &config.DecodeOpts{
			Interpolate:        true,
			InterpolateContext: &c.ctx,
		}, raws...)
		if err != nil {
			return nil, nil, err
		}
	}

	errs := new(packer.MultiError)
	var warnings []string
	errs = packer.MultiErrorAppend(errs, c.Config.Prepare(&c.ctx)...)

	if c.VCenterServer == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vCenter hostname is required"))
	}
	if c.Username == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Username is required"))
	}
	if c.Password == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Password is required"))
	}
	if c.Template == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Template name is required"))
	}
	if c.VMName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target VM name is required"))
	}
	if c.Host == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vSphere host is required"))
	}

	if c.RawShutdownTimeout != "" {
		timeout, err := time.ParseDuration(c.RawShutdownTimeout)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
		}
		c.ShutdownTimeout = timeout
	} else {
		c.ShutdownTimeout = 5 * time.Minute
	}

	if len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

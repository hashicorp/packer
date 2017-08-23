package main

import (
	"github.com/mitchellh/multistep"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type ConnectConfig struct {
	VCenterServer      string `mapstructure:"vcenter_server"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	InsecureConnection bool   `mapstructure:"insecure_connection"`
	Datacenter         string `mapstructure:"datacenter"`
}

func (c *ConnectConfig) Prepare() []error {
	var errs []error

	if c.VCenterServer == "" {
		errs = append(errs, fmt.Errorf("vCenter hostname is required"))
	}
	if c.Username == "" {
		errs = append(errs, fmt.Errorf("Username is required"))
	}
	if c.Password == "" {
		errs = append(errs, fmt.Errorf("Password is required"))
	}

	return errs
}

type StepConnect struct {
	config *ConnectConfig
}

func (s *StepConnect) Run(state multistep.StateBag) multistep.StepAction {
	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      s.config.VCenterServer,
		Username:           s.config.Username,
		Password:           s.config.Password,
		InsecureConnection: s.config.InsecureConnection,
		Datacenter:         s.config.Datacenter,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("driver", d)

	return multistep.ActionContinue
}

func (s *StepConnect) Cleanup(multistep.StateBag) {}

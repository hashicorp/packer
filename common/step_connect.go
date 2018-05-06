package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"context"
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
		errs = append(errs, fmt.Errorf("'vcenter_server' is required"))
	}
	if c.Username == "" {
		errs = append(errs, fmt.Errorf("'username' is required"))
	}
	if c.Password == "" {
		errs = append(errs, fmt.Errorf("'password' is required"))
	}

	return errs
}

type StepConnect struct {
	Config *ConnectConfig
}

func (s *StepConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	d, err := driver.NewDriver(ctx, &driver.ConnectConfig{
		VCenterServer:      s.Config.VCenterServer,
		Username:           s.Config.Username,
		Password:           s.Config.Password,
		InsecureConnection: s.Config.InsecureConnection,
		Datacenter:         s.Config.Datacenter,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("driver", d)

	return multistep.ActionContinue
}

func (s *StepConnect) Cleanup(multistep.StateBag) {}

//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ConnectConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
)

type ConnectConfig struct {
	// vCenter server hostname.
	VCenterServer string `mapstructure:"vcenter_server"`
	// vSphere username.
	Username string `mapstructure:"username"`
	// vSphere password.
	Password string `mapstructure:"password"`
	// Do not validate vCenter server's TLS certificate. Defaults to `false`.
	InsecureConnection bool `mapstructure:"insecure_connection"`
	// VMware datacenter name. Required if there is more than one datacenter in vCenter.
	Datacenter string `mapstructure:"datacenter"`
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

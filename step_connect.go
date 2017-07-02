package main

import (
	"github.com/mitchellh/multistep"
	"context"
	"fmt"
	"net/url"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type ConnectConfig struct {
	VCenterServer      string `mapstructure:"vcenter_server"`
	Datacenter         string `mapstructure:"datacenter"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	InsecureConnection bool   `mapstructure:"insecure_connection"`
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
	ctx := state.Get("ctx").(context.Context)

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", s.config.VCenterServer))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	vcenter_url.User = url.UserPassword(s.config.Username, s.config.Password)
	client, err := govmomi.NewClient(ctx, vcenter_url, s.config.InsecureConnection)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("client", client)

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, s.config.Datacenter)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	finder.SetDatacenter(datacenter)
	state.Put("finder", finder)
	state.Put("datacenter", datacenter)

	return multistep.ActionContinue
}

func (s *StepConnect) Cleanup(multistep.StateBag) {}

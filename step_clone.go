package main

import (
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type CloneConfig struct {
	Template     string `mapstructure:"template"`
	VMName       string `mapstructure:"vm_name"`
	Folder       string `mapstructure:"folder"`
	Host         string `mapstructure:"host"`
	ResourcePool string `mapstructure:"resource_pool"`
	Datastore    string `mapstructure:"datastore"`
	LinkedClone  bool   `mapstructure:"linked_clone"`
}

func (c *CloneConfig) Prepare() []error {
	var errs []error

	if c.Template == "" {
		errs = append(errs, fmt.Errorf("Template name is required"))
	}
	if c.VMName == "" {
		errs = append(errs, fmt.Errorf("Target VM name is required"))
	}
	if c.Host == "" {
		errs = append(errs, fmt.Errorf("vSphere host is required"))
	}

	return errs
}

type StepCloneVM struct {
	config *CloneConfig
}

func (s *StepCloneVM) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	ui.Say("Cloning VM...")

	vm, err := d.CloneVM(&driver.CloneConfig{
		Template:     s.config.Template,
		VMName:       s.config.VMName,
		Folder:       s.config.Folder,
		Host:         s.config.Host,
		ResourcePool: s.config.ResourcePool,
		Datastore:    s.config.Datastore,
		LinkedClone:  s.config.LinkedClone,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	if vm, ok := state.GetOk("vm"); ok {
		ui := state.Get("ui").(packer.Ui)
		d := state.Get("driver").(*driver.Driver)

		ui.Say("Destroying VM...")

		err := d.DestroyVM(vm.(*object.VirtualMachine))
		if err != nil {
			ui.Error(err.Error())
		}
	}
}

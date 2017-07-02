package main

import (
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
	"github.com/hashicorp/packer/packer"
	"fmt"
)

type CloneConfig struct {
	Template     string `mapstructure:"template"`
	FolderName   string `mapstructure:"folder"`
	VMName       string `mapstructure:"vm_name"`
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
	d := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Cloning VM...")

	vm, err := d.cloneVM(s.config)
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
		d := state.Get("driver").(Driver)
		ui := state.Get("ui").(packer.Ui)

		ui.Say("Destroying VM...")

		err := d.destroyVM(vm.(*object.VirtualMachine))
		if err != nil {
			ui.Error(err.Error())
		}
	}
}

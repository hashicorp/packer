package clone

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
)

type CloneConfig struct {
	Template     string `mapstructure:"template"`
	common.VMConfig     `mapstructure:",squash"`
	LinkedClone  bool   `mapstructure:"linked_clone"`
}

func (c *CloneConfig) Prepare() []error {
	errs := c.VMConfig.Prepare()

	if c.Template == "" {
		errs = append(errs, fmt.Errorf("Template name is required"))
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

	template, err := d.FindVM(s.config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vm, err := template.Clone(&driver.CloneConfig{
		Name:         s.config.VMName,
		Folder:       s.config.Folder,
		Cluster:	  s.config.Cluster,
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

	ui := state.Get("ui").(packer.Ui)

	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(*driver.VirtualMachine)

	ui.Say("Destroying VM...")

	err := vm.Destroy()
	if err != nil {
		ui.Error(err.Error())
	}
}

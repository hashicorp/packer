package clone

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type CloneConfig struct {
	Template    string `mapstructure:"template"`
	DiskSize    int64  `mapstructure:"disk_size"`
	LinkedClone bool   `mapstructure:"linked_clone"`
}

func (c *CloneConfig) Prepare() []error {
	var errs []error

	if c.Template == "" {
		errs = append(errs, fmt.Errorf("'template' is required"))
	}

	if c.LinkedClone == true && c.DiskSize != 0 {
		errs = append(errs, fmt.Errorf("'linked_clone' and 'disk_size' cannot be used together"))
	}

	return errs
}

type StepCloneVM struct {
	Config   *CloneConfig
	Location *common.LocationConfig
}

func (s *StepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	ui.Say("Cloning VM...")

	template, err := d.FindVM(s.Config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vm, err := template.Clone(ctx, &driver.CloneConfig{
		Name:         s.Location.VMName,
		Folder:       s.Location.Folder,
		Cluster:      s.Location.Cluster,
		Host:         s.Location.Host,
		ResourcePool: s.Location.ResourcePool,
		Datastore:    s.Location.Datastore,
		LinkedClone:  s.Config.LinkedClone,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if vm == nil {
		return multistep.ActionHalt
	}
	state.Put("vm", vm)

	if s.Config.DiskSize > 0 {
		err = vm.ResizeDisk(s.Config.DiskSize)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

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

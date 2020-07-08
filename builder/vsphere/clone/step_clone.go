//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type CloneConfig,vAppConfig

package clone

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type vAppConfig struct {
	// TODO docs @sylviamoss
	Properties map[string]string `mapstructure:"properties"`
}

type CloneConfig struct {
	// Name of source VM. Path is optional.
	Template string `mapstructure:"template"`
	// The size of the disk in MB.
	DiskSize int64 `mapstructure:"disk_size"`
	// Create VM as a linked clone from latest snapshot. Defaults to `false`.
	LinkedClone bool `mapstructure:"linked_clone"`
	// Set the network in which the VM will be connected to. If no network is specified, `host`
	// must be specified to allow Packer to look for the available network.
	Network string `mapstructure:"network"`
	// VM notes.
	Notes string `mapstructure:"notes"`

	// TODO docs @sylviamoss
	VAppConfig vAppConfig `mapstructure:"vapp"`
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
	Force    bool
}

func (s *StepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)
	vmPath := path.Join(s.Location.Folder, s.Location.VMName)

	err := d.PreCleanVM(ui, vmPath, s.Force)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Cloning VM...")
	template, err := d.FindVM(s.Config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vm, err := template.Clone(ctx, &driver.CloneConfig{
		Name:           s.Location.VMName,
		Folder:         s.Location.Folder,
		Cluster:        s.Location.Cluster,
		Host:           s.Location.Host,
		ResourcePool:   s.Location.ResourcePool,
		Datastore:      s.Location.Datastore,
		LinkedClone:    s.Config.LinkedClone,
		Network:        s.Config.Network,
		Annotation:     s.Config.Notes,
		VAppProperties: s.Config.VAppConfig.Properties,
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

//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type CloneConfig

package clone

import (
	"context"
	"fmt"
	"path"

	"github.com/vmware/govmomi/find"

	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type CloneConfig struct {
	// Name of source VM. Path is optional.
	Template string `mapstructure:"template"`
	// The size of the disk in MB.
	DiskSize int64 `mapstructure:"disk_size"`
	// Create VM as a linked clone from latest snapshot. Defaults to `false`.
	LinkedClone bool `mapstructure:"linked_clone"`
	// Set network VM will be connected to.
	Network string `mapstructure:"network"`
	// VM notes.
	Notes string `mapstructure:"notes"`
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

	vm, err := d.FindVM(vmPath)
	if err != nil {
		if _, ok := err.(*find.NotFoundError); !ok {
			state.Put("error", fmt.Errorf("error looking up old vm: %v", err))
			return multistep.ActionHalt
		}
	} else {
		if s.Force {
			ui.Say(fmt.Sprintf("the vm/template %s already exists, but deleting it due to -force flag", vmPath))

			// power off just in case it is still on
			vm.PowerOff()

			err := vm.Destroy()
			if err != nil {
				state.Put("error", fmt.Errorf("error destroying %s: %v", vmPath, err))
			}
		} else {
			state.Put("error", fmt.Errorf("%s already exists, you can use -force flag to destroy it: %v", vmPath, err))
			return multistep.ActionHalt
		}
	}

	ui.Say("Cloning VM...")
	template, err := d.FindVM(s.Config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vm, err = template.Clone(ctx, &driver.CloneConfig{
		Name:         s.Location.VMName,
		Folder:       s.Location.Folder,
		Cluster:      s.Location.Cluster,
		Host:         s.Location.Host,
		ResourcePool: s.Location.ResourcePool,
		Datastore:    s.Location.Datastore,
		LinkedClone:  s.Config.LinkedClone,
		Network:      s.Config.Network,
		Annotation:   s.Config.Notes,
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

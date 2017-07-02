package main

import (
	"context"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/object"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/vmware/govmomi/vim25/mo"
	"errors"
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

type CloneParameters struct {
	ctx          context.Context
	vmSrc        *object.VirtualMachine
	vmName       string
	folder       *object.Folder
	resourcePool *object.ResourcePool
	datastore    *object.Datastore
	linkedClone  bool
}

type StepCloneVM struct {
	config *CloneConfig
}

func (s *StepCloneVM) Run(state multistep.StateBag) multistep.StepAction {
	d := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmSrc, err := d.finder.VirtualMachine(d.ctx, s.config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("vmSrc", vmSrc)

	ui.Say("Cloning VM...")

	folder, err := d.finder.FolderOrDefault(d.ctx, fmt.Sprintf("/%v/vm/%v", d.datacenter.Name(), s.config.FolderName))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	pool, err := d.finder.ResourcePoolOrDefault(d.ctx, fmt.Sprintf("/%v/host/%v/Resources/%v", d.datacenter.Name(), s.config.Host, s.config.ResourcePool))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	var datastore *object.Datastore
	if s.config.Datastore != "" {
		datastore, err = d.finder.Datastore(d.ctx, s.config.Datastore)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	vm, err := cloneVM(&CloneParameters{
		ctx:          d.ctx,
		vmSrc:        vmSrc,
		vmName:       s.config.VMName,
		folder:       folder,
		resourcePool: pool,
		datastore:    datastore,
		linkedClone:  s.config.LinkedClone,
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
		d := state.Get("driver").(Driver)
		ui := state.Get("ui").(packer.Ui)

		ui.Say("Destroying VM...")

		task, err := vm.(*object.VirtualMachine).Destroy(d.ctx)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		_, err = task.WaitForResult(d.ctx, nil)
		if err != nil {
			ui.Error(err.Error())
			return
		}
	}
}

func cloneVM(params *CloneParameters) (vm *object.VirtualMachine, err error) {
	vm = nil
	err = nil
	poolRef := params.resourcePool.Reference()

	// Creating specs for cloning
	relocateSpec := types.VirtualMachineRelocateSpec{
		Pool: &(poolRef),
	}
	if params.datastore != nil {
		datastoreRef := params.datastore.Reference()
		relocateSpec.Datastore = &datastoreRef
	}
	if params.linkedClone == true {
		relocateSpec.DiskMoveType = "createNewChildDiskBacking"
	}

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: relocateSpec,
		PowerOn:  false,
	}
	if params.linkedClone == true {
		var vmImage mo.VirtualMachine
		err = params.vmSrc.Properties(params.ctx, params.vmSrc.Reference(), []string{"snapshot"}, &vmImage)
		if err != nil {
			err = fmt.Errorf("Error reading base VM properties: %s", err)
			return
		}
		if vmImage.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but image VM has no snapshots")
			return
		}
		cloneSpec.Snapshot = vmImage.Snapshot.CurrentSnapshot
	}

	// Cloning itself
	task, err := params.vmSrc.Clone(params.ctx, params.folder, params.vmName, cloneSpec)
	if err != nil {
		return
	}

	info, err := task.WaitForResult(params.ctx, nil)
	if err != nil {
		return
	}

	vm = object.NewVirtualMachine(params.vmSrc.Client(), info.Result.(types.ManagedObjectReference))
	return vm, nil
}

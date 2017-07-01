package main

import (
	"context"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/object"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/find"
	"fmt"
	"github.com/vmware/govmomi/vim25/mo"
	"errors"
)

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
	config  *Config
	success bool
}

func (s *StepCloneVM) Run(state multistep.StateBag) multistep.StepAction {
	ctx := state.Get("ctx").(context.Context)
	finder := state.Get("finder").(*find.Finder)
	datacenter := state.Get("datacenter").(*object.Datacenter)
	vmSrc := state.Get("vmSrc").(*object.VirtualMachine)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start cloning...")

	// Get folder
	folder, err := finder.FolderOrDefault(ctx, fmt.Sprintf("/%v/vm/%v", datacenter.Name(), s.config.FolderName))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Get resource pool
	pool, err := finder.ResourcePoolOrDefault(ctx, fmt.Sprintf("/%v/host/%v/Resources/%v", datacenter.Name(), s.config.Host, s.config.ResourcePool))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Get datastore
	var datastore *object.Datastore = nil
	if s.config.Datastore != "" {
		datastore, err = finder.Datastore(ctx, s.config.Datastore)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	vm, err := cloneVM(&CloneParameters{
		ctx:          ctx,
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
	s.success = true
	return multistep.ActionContinue
}

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
	if !s.success {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		vm := state.Get("vm").(*object.VirtualMachine)
		ctx := state.Get("ctx").(context.Context)
		ui := state.Get("ui").(packer.Ui)

		ui.Say("destroying VM...")

		task, err := vm.Destroy(ctx)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		_, err = task.WaitForResult(ctx, nil)
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

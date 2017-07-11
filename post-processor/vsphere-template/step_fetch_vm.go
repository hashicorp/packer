package vsphere_template

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
)

type StepFetchVm struct {
	VMName string
	Source string
}

func (s *StepFetchVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ctx := state.Get("context").(context.Context)
	f := state.Get("finder").(*find.Finder)

	ui.Say("Fetching VM...")

	if err := avoidOrphaned(ctx, f, s.VMName); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	path := strings.Split(s.Source, "/vmfs/volumes/")[1]
	i := strings.Index(path, "/")
	storage := path[:i]
	vmx := path[i:]

	ds, err := f.DatastoreOrDefault(ctx, storage)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	folder, err := f.DefaultFolder(ctx)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	pool, err := f.DefaultResourcePool(ctx)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	task, err := folder.RegisterVM(ctx, ds.Path(vmx), s.VMName, false, pool, nil)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err = task.Wait(ctx); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vm, err := f.VirtualMachine(ctx, s.VMName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

// When ESXI remove the VM, vSphere keep the VM as orphaned
func avoidOrphaned(ctx context.Context, f *find.Finder, vm_name string) error {
	vm, err := f.VirtualMachine(ctx, vm_name)
	if err != nil {
		return err
	}
	return vm.Unregister(ctx)
}

func (s *StepFetchVm) Cleanup(multistep.StateBag) {}

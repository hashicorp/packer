package vsphere_template

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
)

type stepFetchVm struct {
	VMName string
	Source string
}

func (s *stepFetchVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	f := state.Get("finder").(*find.Finder)

	ui.Say("Fetching VM...")

	if err := avoidOrphaned(f, s.VMName); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	path := strings.Split(s.Source, "/vmfs/volumes/")[1]
	i := strings.Index(path, "/")
	storage := path[:i]
	vmx := path[i:]

	ds, err := f.DatastoreOrDefault(context.Background(), storage)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	folder, err := f.DefaultFolder(context.Background())
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	pool, err := f.DefaultResourcePool(context.Background())
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	task, err := folder.RegisterVM(context.Background(), ds.Path(vmx), s.VMName, false, pool, nil)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err = task.Wait(context.Background()); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vm, err := f.VirtualMachine(context.Background(), s.VMName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

// When ESXi remove the VM, vSphere keep the VM as orphaned
func avoidOrphaned(f *find.Finder, vm_name string) error {
	vm, err := f.VirtualMachine(context.Background(), vm_name)
	if err != nil {
		return err
	}
	return vm.Unregister(context.Background())
}

func (s *stepFetchVm) Cleanup(multistep.StateBag) {}

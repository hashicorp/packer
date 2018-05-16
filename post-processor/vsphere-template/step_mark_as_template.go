package vsphere_template

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

type stepMarkAsTemplate struct {
	VMName string
	Source string
}

func (s *stepMarkAsTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cli := state.Get("client").(*govmomi.Client)
	folder := state.Get("folder").(*object.Folder)
	dcPath := state.Get("dcPath").(string)

	ui.Message("Marking as a template...")

	vm, err := findRuntimeVM(cli, dcPath, s.VMName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	host, err := vm.HostSystem(context.Background())
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := vm.Unregister(context.Background()); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	source := strings.Split(s.Source, "/vmfs/volumes/")[1]
	i := strings.Index(source, "/")

	path := (&object.DatastorePath{
		Datastore: source[:i],
		Path:      source[i:],
	}).String()

	if err := unregisterPreviousVM(cli, folder, s.VMName); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	task, err := folder.RegisterVM(context.Background(), path, s.VMName, true, nil, host)
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

	return multistep.ActionContinue
}

// We will use the virtual machine created by vmware-iso builder
func findRuntimeVM(cli *govmomi.Client, dcPath, name string) (*object.VirtualMachine, error) {
	si := object.NewSearchIndex(cli.Client)
	fullPath := path.Join(dcPath, "vm", "Discovered virtual machine", name)

	ref, err := si.FindByInventoryPath(context.Background(), fullPath)
	if err != nil {
		return nil, err
	}

	if ref == nil {
		return nil, fmt.Errorf("VM at path %s not found", fullPath)
	}

	return ref.(*object.VirtualMachine), nil
}

// If in the target folder a virtual machine or template already exists
// it will be removed to maintain consistency
func unregisterPreviousVM(cli *govmomi.Client, folder *object.Folder, name string) error {
	si := object.NewSearchIndex(cli.Client)
	fullPath := path.Join(folder.InventoryPath, name)

	ref, err := si.FindByInventoryPath(context.Background(), fullPath)
	if err != nil {
		return err
	}

	if ref != nil {
		if vm, ok := ref.(*object.VirtualMachine); ok {
			return vm.Unregister(context.Background())
		} else {
			return fmt.Errorf("an object name '%v' already exists", name)
		}

	}

	return nil
}

func (s *stepMarkAsTemplate) Cleanup(multistep.StateBag) {}

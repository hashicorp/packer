package vsphere_template

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/vsphere"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type stepMarkAsTemplate struct {
	VMName       string
	RemoteFolder string
}

func NewStepMarkAsTemplate(artifact packer.Artifact) *stepMarkAsTemplate {
	remoteFolder := "Discovered virtual machine"
	vmname := artifact.Id()

	if artifact.BuilderId() == vsphere.BuilderId {
		id := strings.Split(artifact.Id(), "::")
		remoteFolder = id[1]
		vmname = id[2]
	}

	return &stepMarkAsTemplate{
		VMName:       vmname,
		RemoteFolder: remoteFolder,
	}
}

func (s *stepMarkAsTemplate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cli := state.Get("client").(*govmomi.Client)
	folder := state.Get("folder").(*object.Folder)
	dcPath := state.Get("dcPath").(string)

	ui.Message("Marking as a template...")

	vm, err := findRuntimeVM(cli, dcPath, s.VMName, s.RemoteFolder)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dsPath, err := datastorePath(vm)
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

	if err := unregisterPreviousVM(cli, folder, s.VMName); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = common.Retry(2, 10, 30, func(i uint) (bool, error) {
		ui.Message(fmt.Sprintf("Registering VM, attempt %d", i+1))

		task, err := folder.RegisterVM(context.Background(), dsPath.String(), s.VMName, true, nil, host)
		if err == nil {
			// success
			return true, nil
		}
		ui.Message(fmt.Sprintf("Error registering VM; will retry..."+
			"Error: %s", err.Error()))
		// retry
		return false, nil
	})

	if err != nil {
		state.Put("error registering vm")
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

func datastorePath(vm *object.VirtualMachine) (*object.DatastorePath, error) {
	devices, err := vm.Device(context.Background())
	if err != nil {
		return nil, err
	}

	disk := ""
	for _, device := range devices {
		if d, ok := device.(*types.VirtualDisk); ok {
			if b, ok := d.Backing.(types.BaseVirtualDeviceFileBackingInfo); ok {
				disk = b.GetVirtualDeviceFileBackingInfo().FileName
			}
			break
		}
	}

	if disk == "" {
		return nil, fmt.Errorf("disk not found in '%v'", vm.Name())
	}

	re := regexp.MustCompile("\\[(.*?)\\]")

	datastore := re.FindStringSubmatch(disk)[1]
	vmxPath := path.Join("/", path.Dir(strings.Split(disk, " ")[1]), vm.Name()+".vmx")

	return &object.DatastorePath{
		Datastore: datastore,
		Path:      vmxPath,
	}, nil
}

func findRuntimeVM(cli *govmomi.Client, dcPath, name, remoteFolder string) (*object.VirtualMachine, error) {
	si := object.NewSearchIndex(cli.Client)
	fullPath := path.Join(dcPath, "vm", remoteFolder, name)

	ref, err := si.FindByInventoryPath(context.Background(), fullPath)
	if err != nil {
		return nil, err
	}

	if ref == nil {
		return nil, fmt.Errorf("VM at path %s not found", fullPath)
	}

	vm := ref.(*object.VirtualMachine)
	if vm.InventoryPath == "" {
		vm.SetInventoryPath(fullPath)
	}

	return vm, nil
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

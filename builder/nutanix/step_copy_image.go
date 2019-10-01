package nutanix

import (
	"context"
	"errors"
	"time"

	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCopyImage struct {
	Config *Config
}

func (s *stepCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmUUID := state.Get("vmUUID").(string)

	ui.Say("Saving VM for uuid: " + vmUUID)
	ui.Message("Retrieving VM status")

	d := NewDriver(&s.Config.NutanixCluster, state)
	vmResponse, err := d.RetrieveReadyVM(ctx, 1*time.Minute)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message("Initiatiating save VM DISK task.")
	// Choose disk to replicate - looking for first "DISK"
	var diskToCopy string
	for i := range vmResponse.Spec.Resources.DiskList {
		if *(vmResponse.Spec.Resources.DiskList)[i].DeviceProperties.DeviceType == "DISK" {
			diskToCopy = *(vmResponse.Spec.Resources.DiskList)[i].UUID
			ui.Message("Found DISK to copy: " + diskToCopy)
			break
		}
	}
	if diskToCopy == "" {
		err := errors.New("No DISK was found to save, halting build")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	vmReqKind := "vm_disk"
	diskReqKind := "image"
	vmIntentInput := &v3.ImageIntentInput{
		Spec: &v3.Image{
			Name: &s.Config.NewImageName,
			Resources: &v3.ImageResources{
				DataSourceReference: &v3.Reference{
					Kind: &vmReqKind,
					UUID: &diskToCopy,
				},
			},
		},
		Metadata: &v3.Metadata{
			Kind: &diskReqKind,
		},
	}
	vmResponse, err = d.SaveVMDisk(ctx, vmIntentInput)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	var taskUUID string
	if *vmResponse.Status.State == "PENDING" {
		taskUUID = vmResponse.Status.ExecutionContext.TaskUUID.(string)
		ui.Message("Copy vm_disk task submitted, waiting for completion: " + taskUUID)
		taskResponse, err := d.RetrieveTask(ctx, taskUUID)
		if err != nil {
			ui.Error("Unexpected Nutanix Task status: " + err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Message("Successfully saved vm disk: " + *(taskResponse.EntityReferenceList)[0].UUID)
		state.Put("vm_disk_uuid", (taskResponse.EntityReferenceList)[0].UUID)
	} else {
		err := errors.New("Unexpected Nutanix Task status: " + *vmResponse.Status.State)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {
}

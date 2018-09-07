package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepAttachVolume struct {
	index           int
	volumeName      string
	instanceInfoKey string
}

func (s *stepAttachVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*compute.ComputeClient)
	ui := state.Get("ui").(packer.Ui)
	instanceInfo := state.Get(s.instanceInfoKey).(*compute.InstanceInfo)

	saClient := client.StorageAttachments()
	saInput := &compute.CreateStorageAttachmentInput{
		Index:             s.index,
		InstanceName:      instanceInfo.Name + "/" + instanceInfo.ID,
		StorageVolumeName: s.volumeName,
	}

	sa, err := saClient.CreateStorageAttachment(saInput)
	if err != nil {
		err = fmt.Errorf("Problem attaching master volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put(s.instanceInfoKey+"/attachment", sa)

	ui.Message("Volume to master attached to builder instance")
	return multistep.ActionContinue
}

func (s *stepAttachVolume) Cleanup(state multistep.StateBag) {
	sa, ok := state.GetOk(s.instanceInfoKey + "/attachment")
	if !ok {
		return
	}
	client := state.Get("client").(*compute.ComputeClient)
	ui := state.Get("ui").(packer.Ui)

	saClient := client.StorageAttachments()
	saI := &compute.DeleteStorageAttachmentInput{
		Name: sa.(*compute.StorageAttachmentInfo).Name,
	}

	if err := saClient.DeleteStorageAttachment(saI); err != nil {
		err = fmt.Errorf("Problem detaching storage volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
}

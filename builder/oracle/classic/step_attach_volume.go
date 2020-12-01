package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepAttachVolume struct {
	Index           int
	VolumeName      string
	InstanceInfoKey string
}

func (s *stepAttachVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*compute.Client)
	ui := state.Get("ui").(packersdk.Ui)
	instanceInfo := state.Get(s.InstanceInfoKey).(*compute.InstanceInfo)

	saClient := client.StorageAttachments()
	saInput := &compute.CreateStorageAttachmentInput{
		Index:             s.Index,
		InstanceName:      instanceInfo.Name + "/" + instanceInfo.ID,
		StorageVolumeName: s.VolumeName,
	}

	sa, err := saClient.CreateStorageAttachment(saInput)
	if err != nil {
		err = fmt.Errorf("Problem attaching master volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put(s.InstanceInfoKey+"/attachment", sa)

	ui.Message("Volume attached to instance.")
	return multistep.ActionContinue
}

func (s *stepAttachVolume) Cleanup(state multistep.StateBag) {
	sa, ok := state.GetOk(s.InstanceInfoKey + "/attachment")
	if !ok {
		return
	}
	client := state.Get("client").(*compute.Client)
	ui := state.Get("ui").(packersdk.Ui)

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

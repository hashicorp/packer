package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreatePersistentVolume struct {
	volumeSize      string
	volumeName      string
	bootable        bool
	sourceImageList string
}

func (s *stepCreatePersistentVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*compute.ComputeClient)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Volume...")

	c := &compute.CreateStorageVolumeInput{
		Name:       s.volumeName,
		Size:       s.volumeSize,
		ImageList:  s.sourceImageList,
		Properties: []string{"/oracle/public/storage/default"},
		Bootable:   s.bootable,
	}

	sc := client.StorageVolumes()
	cc, err := sc.CreateStorageVolume(c)

	if err != nil {
		err = fmt.Errorf("Error creating persistent storage volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	//TODO: wait to become available

	ui.Message(fmt.Sprintf("Created volume: %s", cc.Name))
	return multistep.ActionContinue
}

func (s *stepCreatePersistentVolume) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*compute.ComputeClient)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Cleaning up Volume...")

	c := &compute.DeleteStorageVolumeInput{
		Name: s.volumeName,
	}

	sc := client.StorageVolumes()

	if err := sc.DeleteStorageVolume(c); err != nil {
		ui.Error(fmt.Sprintf("Error cleaning up persistent storage volume: %s", err))
		return
	}

	ui.Message(fmt.Sprintf("Deleted volume: %s", s.volumeName))
}

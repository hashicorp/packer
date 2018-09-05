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
	latencyStorage  bool
	sourceImageList string
}

func (s *stepCreatePersistentVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*compute.ComputeClient)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Volume...")

	var properties string
	if s.latencyStorage {
		properties = "/oracle/public/storage/latency"
	} else {
		properties = "/oracle/public/storage/default"
	}

	c := &compute.CreateStorageVolumeInput{
		Name:       s.volumeName,
		Size:       s.volumeSize,
		Properties: []string{properties},
		ImageList:  s.sourceImageList,
		Bootable:   true,
	}

	sc := client.StorageVolumes()
	cc, err := sc.CreateStorageVolume(c)

	if err != nil {
		err = fmt.Errorf("Error creating persistent volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	//TODO: wait to become available

	ui.Message(fmt.Sprintf("Created volume: %s", cc.Name))
	return multistep.ActionContinue
}

func (s *stepCreatePersistentVolume) Cleanup(state multistep.StateBag) {
}

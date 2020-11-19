package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreatePersistentVolume struct {
	VolumeSize     string
	VolumeName     string
	Bootable       bool
	ImageList      string
	ImageListEntry int
}

func (s *stepCreatePersistentVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*compute.Client)

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Volume...")

	c := &compute.CreateStorageVolumeInput{
		Name:           s.VolumeName,
		Size:           s.VolumeSize,
		ImageList:      s.ImageList,
		ImageListEntry: s.ImageListEntry,
		Properties:     []string{"/oracle/public/storage/default"},
		Bootable:       s.Bootable,
	}

	sc := client.StorageVolumes()
	cc, err := sc.CreateStorageVolume(c)

	if err != nil {
		err = fmt.Errorf("Error creating persistent storage volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Created volume: %s", cc.Name))
	return multistep.ActionContinue
}

func (s *stepCreatePersistentVolume) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*compute.Client)

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Cleaning up Volume...")

	c := &compute.DeleteStorageVolumeInput{
		Name: s.VolumeName,
	}

	sc := client.StorageVolumes()

	if err := sc.DeleteStorageVolume(c); err != nil {
		ui.Error(fmt.Sprintf("Error cleaning up persistent storage volume: %s", err))
		return
	}

	ui.Message(fmt.Sprintf("Deleted volume: %s", s.VolumeName))
}

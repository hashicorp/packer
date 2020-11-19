package chroot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

var _ multistep.Step = &StepAttachDisk{}

type StepAttachDisk struct {
	attached bool
}

func (s *StepAttachDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	diskset := state.Get(stateBagKey_Diskset).(Diskset)
	diskResourceID := diskset.OS().String()

	ui.Say(fmt.Sprintf("Attaching disk '%s'", diskResourceID))

	da := NewDiskAttacher(azcli)
	lun, err := da.AttachDisk(ctx, diskResourceID)
	if err != nil {
		log.Printf("StepAttachDisk.Run: error: %+v", err)
		err := fmt.Errorf(
			"error attaching disk '%s': %v", diskResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Disk attached, waiting for device to show up")
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3) // in case is not configured correctly
	defer cancel()
	device, err := da.WaitForDevice(ctx, lun)
	if err != nil {
		log.Printf("StepAttachDisk.Run: error: %+v", err)
		err := fmt.Errorf(
			"error attaching disk '%s': %v", diskResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Disk available at %q", device))
	s.attached = true
	state.Put("device", device)
	state.Put("attach_cleanup", s)
	return multistep.ActionContinue
}

func (s *StepAttachDisk) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepAttachDisk) CleanupFunc(state multistep.StateBag) error {

	if s.attached {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packersdk.Ui)
		diskset := state.Get(stateBagKey_Diskset).(Diskset)
		diskResourceID := diskset.OS().String()

		ui.Say(fmt.Sprintf("Detaching disk '%s'", diskResourceID))

		da := NewDiskAttacher(azcli)
		err := da.DetachDisk(context.Background(), diskResourceID)
		if err != nil {
			return fmt.Errorf("error detaching %q: %v", diskResourceID, err)
		}
		s.attached = false
	}

	return nil
}

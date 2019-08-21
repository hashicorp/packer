package chroot

import (
	"context"
	"fmt"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepLinkVolume attaches the previously created volume to an
// available device location.
//
// Produces:
//   device string - The location where the volume was attached.
//   attach_cleanup CleanupFunc
type StepLinkVolume struct {
	attached bool
	volumeId string
}

func (s *StepLinkVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	device := state.Get("device").(string)
	vm := state.Get("vm").(oapi.Vm)
	ui := state.Get("ui").(packer.Ui)
	volumeId := state.Get("volume_id").(string)

	// For the API call, it expects "sd" prefixed devices.
	//linkVolume := strings.Replace(device, "/xvd", "/sd", 1)
	linkVolume := device

	ui.Say(fmt.Sprintf("Attaching the root volume to %s", linkVolume))
	_, err := oapiconn.POST_LinkVolume(oapi.LinkVolumeRequest{
		VmId:       vm.VmId,
		VolumeId:   volumeId,
		DeviceName: linkVolume,
	})

	if err != nil {
		err := fmt.Errorf("Error attaching volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Mark that we attached it so we can detach it later
	s.attached = true
	s.volumeId = volumeId

	// Wait for the volume to become attached
	err = osccommon.WaitUntilVolumeIsLinked(oapiconn, s.volumeId)
	if err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("attach_cleanup", s)
	return multistep.ActionContinue
}

func (s *StepLinkVolume) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepLinkVolume) CleanupFunc(state multistep.StateBag) error {
	if !s.attached {
		return nil
	}

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Detaching BSU volume...")
	_, err := oapiconn.POST_UnlinkVolume(oapi.UnlinkVolumeRequest{VolumeId: s.volumeId})
	if err != nil {
		return fmt.Errorf("Error detaching BSU volume: %s", err)
	}

	s.attached = false

	// Wait for the volume to detach
	err = osccommon.WaitUntilVolumeIsUnlinked(oapiconn, s.volumeId)
	if err != nil {
		return fmt.Errorf("Error waiting for volume: %s", err)
	}

	return nil
}

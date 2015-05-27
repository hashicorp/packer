package common

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type guestAdditionsPathTemplate struct {
	Version string
}

// This step uploads the guest additions ISO to the VM.
type StepUploadGuestAdditions struct {
	GuestAdditionsMode string
	GuestAdditionsPath string
	Ctx                interpolate.Context
}

func (s *StepUploadGuestAdditions) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// If we're attaching then don't do this, since we attached.
	if s.GuestAdditionsMode != GuestAdditionsModeUpload {
		log.Println("Not uploading guest additions since mode is not upload")
		return multistep.ActionContinue
	}

	// Get the guest additions path since we're doing it
	guestAdditionsPath := state.Get("guest_additions_path").(string)

	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for guest additions upload: %s", err))
		return multistep.ActionHalt
	}

	f, err := os.Open(guestAdditionsPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening guest additions ISO: %s", err))
		return multistep.ActionHalt
	}

	s.Ctx.Data = &guestAdditionsPathTemplate{
		Version: version,
	}

	s.GuestAdditionsPath, err = interpolate.Render(s.GuestAdditionsPath, &s.Ctx)
	if err != nil {
		err := fmt.Errorf("Error preparing guest additions path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading VirtualBox guest additions ISO...")
	if err := comm.Upload(s.GuestAdditionsPath, f, nil); err != nil {
		state.Put("error", fmt.Errorf("Error uploading guest additions: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUploadGuestAdditions) Cleanup(state multistep.StateBag) {}

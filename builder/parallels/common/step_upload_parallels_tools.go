package common

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// This step uploads the Parallels Tools ISO to the virtual machine.
//
// Uses:
//   communicator packer.Communicator
//   parallels_tools_path string
//   ui packer.Ui
//
// Produces:
type toolsPathTemplate struct {
	Flavor string
}

// This step uploads the guest additions ISO to the VM.
type StepUploadParallelsTools struct {
	ParallelsToolsFlavor    string
	ParallelsToolsGuestPath string
	ParallelsToolsMode      string
	Ctx                     interpolate.Context
}

func (s *StepUploadParallelsTools) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)

	// If we're attaching then don't do this, since we attached.
	if s.ParallelsToolsMode != ParallelsToolsModeUpload {
		log.Println("Not uploading Parallels Tools since mode is not upload")
		return multistep.ActionContinue
	}

	// Get the Paralells Tools path on the host machine
	parallelsToolsPath := state.Get("parallels_tools_path").(string)

	f, err := os.Open(parallelsToolsPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening Parallels Tools ISO: %s", err))
		return multistep.ActionHalt
	}
	defer f.Close()

	s.Ctx.Data = &toolsPathTemplate{
		Flavor: s.ParallelsToolsFlavor,
	}

	s.ParallelsToolsGuestPath, err = interpolate.Render(s.ParallelsToolsGuestPath, &s.Ctx)
	if err != nil {
		err := fmt.Errorf("Error preparing Parallels Tools path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading Parallels Tools for '%s' to path: '%s'",
		s.ParallelsToolsFlavor, s.ParallelsToolsGuestPath))
	if err := comm.Upload(s.ParallelsToolsGuestPath, f, nil); err != nil {
		err := fmt.Errorf("Error uploading Parallels Tools: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUploadParallelsTools) Cleanup(state multistep.StateBag) {}

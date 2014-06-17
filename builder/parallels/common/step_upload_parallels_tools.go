package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
)

type toolsPathTemplate struct {
	Version string
}

// This step uploads the guest additions ISO to the VM.
type StepUploadParallelsTools struct {
	ParallelsToolsHostPath  string
	ParallelsToolsGuestPath string
	ParallelsToolsMode      string
	Tpl                     *packer.ConfigTemplate
}

func (s *StepUploadParallelsTools) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// If we're attaching then don't do this, since we attached.
	if s.ParallelsToolsMode != ParallelsToolsModeUpload {
		log.Println("Not uploading Parallels Tools since mode is not upload")
		return multistep.ActionContinue
	}

	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for Parallels Tools upload: %s", err))
		return multistep.ActionHalt
	}

	f, err := os.Open(s.ParallelsToolsHostPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening Parallels Tools ISO: %s", err))
		return multistep.ActionHalt
	}

	tplData := &toolsPathTemplate{
		Version: version,
	}

	s.ParallelsToolsGuestPath, err = s.Tpl.Process(s.ParallelsToolsGuestPath, tplData)
	if err != nil {
		err := fmt.Errorf("Error preparing Parallels Tools path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading Parallels Tools ISO...")
	if err := comm.Upload(s.ParallelsToolsGuestPath, f); err != nil {
		state.Put("error", fmt.Errorf("Error uploading Parallels Tools: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUploadParallelsTools) Cleanup(state multistep.StateBag) {}

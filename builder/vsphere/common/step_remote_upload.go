package common

import (
	"fmt"
	"log"
	"path"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepRemoteUpload uploads something from the state bag to a driver
// (if it can) and stores that new remote path into the state bag.
type StepRemoteUpload struct {
	Key     string
	Message string
}

func (s *StepRemoteUpload) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	filePath, ok := state.Get(s.Key).(string)
	if !ok {
		return multistep.ActionContinue
	}

	ui.Say(s.Message)
	log.Printf("Remote uploading: %s", filePath)
	newPath, err := driver.Upload(filePath, path.Base(filePath))
	if err != nil {
		err := fmt.Errorf("Error uploading file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(s.Key, newPath)
	return multistep.ActionContinue
}

func (s *StepRemoteUpload) Cleanup(state multistep.StateBag) {
}

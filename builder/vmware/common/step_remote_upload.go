package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// stepRemoteUpload uploads some thing from the state bag to a remote driver
// (if it can) and stores that new remote path into the state bag.
type StepRemoteUpload struct {
	Key       string
	Message   string
	DoCleanup bool
	Checksum  string
}

func (s *StepRemoteUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	remote, ok := driver.(RemoteDriver)
	if !ok {
		return multistep.ActionContinue
	}

	path, ok := state.Get(s.Key).(string)
	if !ok {
		return multistep.ActionContinue
	}

	if esx5, ok := remote.(*ESX5Driver); ok {
		remotePath := esx5.CachePath(path)

		if esx5.VerifyChecksum(s.Checksum, remotePath) {
			ui.Say("Remote cache was verified skipping remote upload...")
			state.Put(s.Key, remotePath)
			return multistep.ActionContinue
		}

	}

	ui.Say(s.Message)
	log.Printf("Remote uploading: %s", path)
	newPath, err := remote.UploadISO(path, s.Checksum, ui)
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
	if !s.DoCleanup {
		return
	}

	driver := state.Get("driver").(Driver)

	remote, ok := driver.(RemoteDriver)
	if !ok {
		return
	}

	path, ok := state.Get(s.Key).(string)
	if !ok {
		return
	}

	log.Printf("Cleaning up remote path: %s", path)
	err := remote.RemoveCache(path)
	if err != nil {
		log.Printf("Error cleaning up: %s", err)
	}
}

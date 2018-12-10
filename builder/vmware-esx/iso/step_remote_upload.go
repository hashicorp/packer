package iso

import (
	"context"
	"fmt"
	"log"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// stepRemoteUpload uploads some thing from the state bag to a remote driver
// (if it can) and stores that new remote path into the state bag.
type stepRemoteUpload struct {
	Key       string
	Message   string
	DoCleanup bool
}

func (s *stepRemoteUpload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	// Ensure that the driver is a remote driver (it should always be)
	remote, ok := driver.(vmwcommon.RemoteDriver)
	if !ok {
		return multistep.ActionContinue
	}

	// Grab the key for the cache path
	path, ok := state.Get(s.Key).(string)
	if !ok {
		return multistep.ActionContinue
	}

	// Grab the template configuration and its checksum
	config := state.Get("config").(*Config)
	checksum := config.ISOChecksum
	checksumType := config.ISOChecksumType

	// Verify the checksum of the remote cache so we can skip upload if it matches
	if esx5, ok := remote.(*vmwcommon.ESX5Driver); ok {
		remotePath := esx5.CachePath(path)

		if esx5.VerifyChecksum(checksumType, checksum, remotePath) {
			ui.Say("Remote cache was verified skipping remote upload...")
			state.Put(s.Key, remotePath)
			return multistep.ActionContinue
		}

	// If we're not using the ESX5 checksum, then just notify the user we're
	// going to re-upload anyways. (Matches logic of previous ESX implementation)
	} else {
		ui.Say("Remote driver does not support the ability to verify the checksum. Re-uploading..")
	}

	// Notify the user and log what's going on
	ui.Say(s.Message)
	log.Printf("Remote uploading: %s", path)

	// Proceed to upload the ISO
	newPath, err := remote.UploadISO(path, checksum, checksumType)
	if err != nil {
		err := fmt.Errorf("Error uploading file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put(s.Key, newPath)

	// We should be good to continue to the next step...
	return multistep.ActionContinue
}

func (s *stepRemoteUpload) Cleanup(state multistep.StateBag) {
	if !s.DoCleanup {
		return
	}

	driver := state.Get("driver").(vmwcommon.Driver)

	// Verify that the driver implements RemoteDriver (it should)
	remote, ok := driver.(vmwcommon.RemoteDriver)
	if !ok {
		return
	}

	// Try and grab the path, if not then there's nothing to clean up
	path, ok := state.Get(s.Key).(string)
	if !ok {
		return
	}

	// Remove the remotely cached upload data
	log.Printf("Cleaning up remote path: %s", path)
	err := remote.RemoveCache(path)
	if err != nil {
		log.Printf("Error cleaning up: %s", err)
	}
}

package iso

import (
	"fmt"
	"log"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// stepRemoteUpload uploads some thing from the state bag to a remote driver
// (if it can) and stores that new remote path into the state bag.
type stepRemoteUpload struct {
	Key     string
	Message string

	// Set this to true for skip
	Skip bool
}

func (s *stepRemoteUpload) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	if s.Skip {
		return multistep.ActionContinue
	}

	remote, ok := driver.(RemoteDriver)
	if !ok {
		return multistep.ActionContinue
	}

	path, ok := state.Get(s.Key).(string)
	if !ok {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)
	checksum := config.ISOChecksum
	checksumType := config.ISOChecksumType

	if esx5, ok := remote.(*ESX5Driver); ok {
		remotePath := esx5.cachePath(path)

		if esx5.verifyChecksum(checksumType, checksum, remotePath) {
			ui.Say("Remote cache was verified skipping remote upload...")
			state.Put(s.Key, remotePath)
			return multistep.ActionContinue
		}

	}

	ui.Say(s.Message)
	log.Printf("Remote uploading: %s", path)
	newPath, err := remote.UploadISO(path, checksum, checksumType)
	if err != nil {
		err := fmt.Errorf("Error uploading file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put(s.Key, newPath)

	return multistep.ActionContinue
}

func (s *stepRemoteUpload) Cleanup(state multistep.StateBag) {
}

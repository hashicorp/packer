package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

// stepRemoteUpload uploads some thing from the state bag to a remote driver
// (if it can) and stores that new remote path into the state bag.
type stepRemoteUpload struct {
	Key     string
	Message string
}

func (s *stepRemoteUpload) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

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

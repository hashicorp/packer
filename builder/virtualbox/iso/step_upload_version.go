package iso

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
type stepUploadVersion struct{}

func (s *stepUploadVersion) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	if config.VBoxVersionFile == "" {
		log.Println("VBoxVersionFile is empty. Not uploading.")
		return multistep.ActionContinue
	}

	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for metadata upload: %s", err))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading VirtualBox version info (%s)", version))
	var data bytes.Buffer
	data.WriteString(version)
	if err := comm.Upload(config.VBoxVersionFile, &data); err != nil {
		state.Put("error", fmt.Errorf("Error uploading VirtualBox version: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUploadVersion) Cleanup(state multistep.StateBag) {}

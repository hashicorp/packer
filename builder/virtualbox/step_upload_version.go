package virtualbox

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
type stepUploadVersion struct{}

func (s *stepUploadVersion) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	version, err := driver.Version()
	if err != nil {
		state["error"] = fmt.Errorf("Error reading version for metadata upload: %s", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading VirtualBox version info (%s)", version))
	var data bytes.Buffer
	data.WriteString(version)
	if err := comm.Upload(config.VBoxVersionFile, &data); err != nil {
		state["error"] = fmt.Errorf("Error uploading VirtualBox version: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUploadVersion) Cleanup(state map[string]interface{}) {}

package virtualbox

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"text/template"
)

type guestAdditionsPathTemplate struct {
	Version string
}

// This step uploads the guest additions ISO to the VM.
type stepUploadGuestAdditions struct{}

func (s *stepUploadGuestAdditions) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	guestAdditionsPath := state["guest_additions_path"].(string)
	ui := state["ui"].(packer.Ui)

	version, err := driver.Version()
	if err != nil {
		state["error"] = fmt.Errorf("Error reading version for guest additions upload: %s", err)
		return multistep.ActionHalt
	}

	f, err := os.Open(guestAdditionsPath)
	if err != nil {
		state["error"] = fmt.Errorf("Error opening guest additions ISO: %s", err)
		return multistep.ActionHalt
	}

	tplData := &guestAdditionsPathTemplate{
		Version: version,
	}

	var processedPath bytes.Buffer
	t := template.Must(template.New("path").Parse(config.GuestAdditionsPath))
	t.Execute(&processedPath, tplData)

	ui.Say("Upload VirtualBox guest additions ISO...")
	if err := comm.Upload(processedPath.String(), f); err != nil {
		state["error"] = fmt.Errorf("Error uploading guest additions: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUploadGuestAdditions) Cleanup(state map[string]interface{}) {}

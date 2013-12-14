package googlecompute

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateImage represents a Packer build step that creates GCE machine
// images.
type StepCreateImage int

// Run executes the Packer build step that creates a GCE machine image.
//
// Currently the only way to create a GCE image is to run the gcimagebundle
// command on the running GCE instance.
func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)

	sudoPrefix := ""
	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}

	imageFilename := fmt.Sprintf("%s.tar.gz", config.ImageName)
	imageBundleCmd := "/usr/bin/gcimagebundle -d /dev/sda -o /tmp/"

	ui.Say("Creating image...")
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s%s --output_file_name %s",
		sudoPrefix, imageBundleCmd, imageFilename)
	err := cmd.StartWithUi(comm, ui)
	if err == nil && cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"gcimagebundle exited with non-zero exit status: %d", cmd.ExitStatus)
	}
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image_file_name", filepath.Join("/tmp", imageFilename))
	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {}

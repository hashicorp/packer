package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepUploadImage represents a Packer build step that uploads GCE machine images.
type StepUploadImage int

// Run executes the Packer build step that uploads a GCE machine image.
func (s *StepUploadImage) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	imageFilename := state.Get("image_file_name").(string)
	ui := state.Get("ui").(packer.Ui)

	sudoPrefix := ""
	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}

	ui.Say("Uploading image...")
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s/usr/local/bin/gsutil cp %s gs://%s",
		sudoPrefix, imageFilename, config.BucketName)
	err := cmd.StartWithUi(comm, ui)
	if err == nil && cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"gsutil exited with non-zero exit status: %d", cmd.ExitStatus)
	}
	if err != nil {
		err := fmt.Errorf("Error uploading image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup.
func (s *StepUploadImage) Cleanup(state multistep.StateBag) {}

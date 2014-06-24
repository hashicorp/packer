package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepUpload struct {
}

func (s *stepUpload) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	upload := state.Get("upload").(*Upload)
	artifactFilePath := state.Get("artifactFilePath").(string)
	url := upload.UploadPath

	ui.Say(fmt.Sprintf("Uploading box: %s", artifactFilePath))

	resp, err := client.Upload(artifactFilePath, url)

	if err != nil || (resp.StatusCode != 200) {
		state.Put("error", fmt.Errorf("Error uploading Box: %s", resp.Body))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

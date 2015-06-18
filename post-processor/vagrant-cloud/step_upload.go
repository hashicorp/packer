package vagrantcloud

import (
	"fmt"
	"time"

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
	ui.Message(
		"Depending on your internet connection and the size of the box,\n" +
			"this may take some time")

	var finalErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			ui.Message(fmt.Sprintf("Uploading box, attempt %d", i+1))
		}

		resp, err := client.Upload(artifactFilePath, url)
		if err != nil {
			finalErr = err
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Error: %s", err))
			time.Sleep(10 * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			finalErr = fmt.Errorf("bad HTTP status: %d", resp.StatusCode)
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Status: %d",
				resp.StatusCode))
			time.Sleep(10 * time.Second)
			continue
		}

		finalErr = nil
	}

	if finalErr != nil {
		state.Put("error", finalErr)
		return multistep.ActionHalt
	}

	ui.Message("Box succesfully uploaded")

	return multistep.ActionContinue
}

func (s *stepUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

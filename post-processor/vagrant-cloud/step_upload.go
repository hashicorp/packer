package vagrantcloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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

	err := common.Retry(10, 10, 3, func(i uint) (bool, error) {
		ui.Message(fmt.Sprintf("Uploading box, attempt %d", i+1))

		resp, err := client.Upload(artifactFilePath, url)
		if err != nil {
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Error: %s", err))
			return false, nil
		}
		if resp.StatusCode != 200 {
			log.Printf("bad HTTP status: %d", resp.StatusCode)
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Status: %d",
				resp.StatusCode))
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message("Box successfully uploaded")

	return multistep.ActionContinue
}

func (s *stepUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

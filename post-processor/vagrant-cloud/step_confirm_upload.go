package vagrantcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfirmUpload struct {
}

func (s *stepConfirmUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	upload := state.Get("upload").(*Upload)
	url := upload.CallbackPath

	ui.Say("Confirming direct box upload completion")

	resp, err := client.Callback(url)

	if err != nil || resp.StatusCode != 200 {
		if resp == nil || resp.Body == nil {
			state.Put("error", "No response from server.")
		} else {
			cloudErrors := &VagrantCloudErrors{}
			err = decodeBody(resp, cloudErrors)
			if err != nil {
				ui.Error(fmt.Sprintf("error decoding error response: %s", err))
			}
			state.Put("error", fmt.Errorf("Error preparing upload: %s", cloudErrors.FormatErrors()))
		}
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepConfirmUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

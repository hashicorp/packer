package vagrantcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type Upload struct {
	UploadPath   string `json:"upload_path"`
	CallbackPath string `json:"callback"`
}

type stepPrepareUpload struct {
}

func (s *stepPrepareUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)
	provider := state.Get("provider").(*Provider)
	artifactFilePath := state.Get("artifactFilePath").(string)

	path := fmt.Sprintf("box/%s/version/%v/provider/%s/upload", box.Tag, version.Version, provider.Name)
	if !config.NoDirectUpload {
		path = path + "/direct"
	}
	upload := &Upload{}

	ui.Say(fmt.Sprintf("Preparing upload of box: %s", artifactFilePath))

	resp, err := client.Get(path)

	if err != nil || (resp.StatusCode != 200) {
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

	if err = decodeBody(resp, upload); err != nil {
		state.Put("error", fmt.Errorf("Error parsing upload response: %s", err))
		return multistep.ActionHalt
	}

	// Save the upload details to the state
	state.Put("upload", upload)

	return multistep.ActionContinue
}

func (s *stepPrepareUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

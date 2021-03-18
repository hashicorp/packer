package vagrantcloud

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const VAGRANT_CLOUD_DIRECT_UPLOAD_LIMIT = 5000000000 // Upload limit is 5GB

type Upload struct {
	UploadPath   string `json:"upload_path"`
	CallbackPath string `json:"callback"`
}

type stepPrepareUpload struct {
}

func (s *stepPrepareUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)
	provider := state.Get("provider").(*Provider)
	artifactFilePath := state.Get("artifactFilePath").(string)

	// If direct upload is enabled, the asset size must be <= 5 GB
	if config.NoDirectUpload == false {
		f, err := os.Stat(artifactFilePath)
		if err != nil {
			ui.Error(fmt.Sprintf("error determining size of upload artifact: %s", artifactFilePath))
		}
		if f.Size() > VAGRANT_CLOUD_DIRECT_UPLOAD_LIMIT {
			ui.Say(fmt.Sprintf("Asset %s is larger than the direct upload limit. Setting `NoDirectUpload` to true", artifactFilePath))
			config.NoDirectUpload = true
		}
	}

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

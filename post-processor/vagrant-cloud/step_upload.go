package vagrantcloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"
)

type stepUpload struct {
}

func (s *stepUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	upload := state.Get("upload").(*Upload)
	artifactFilePath := state.Get("artifactFilePath").(string)
	url := upload.UploadPath

	ui.Say(fmt.Sprintf("Uploading box: %s", artifactFilePath))
	ui.Message(
		"Depending on your internet connection and the size of the box,\n" +
			"this may take some time")

	err := retry.Config{
		Tries:      3,
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 10 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		ui.Message(fmt.Sprintf("Uploading box"))

		var err error
		var resp *http.Response

		if config.NoDirectUpload {
			resp, err = client.Upload(artifactFilePath, url)
		} else {
			resp, err = client.DirectUpload(artifactFilePath, url)
		}
		if err != nil {
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Error: %s", err))
			return err
		}
		if resp.StatusCode != 200 {
			err := fmt.Errorf("bad HTTP status: %d", resp.StatusCode)
			log.Print(err)
			ui.Message(fmt.Sprintf(
				"Error uploading box! Will retry in 10 seconds. Status: %d",
				resp.StatusCode))
			return err
		}
		return err
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

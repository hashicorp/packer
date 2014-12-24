package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepVerifyUpload struct {
}

func (s *stepVerifyUpload) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)
	upload := state.Get("upload").(*Upload)
	provider := state.Get("provider").(*Provider)

	path := fmt.Sprintf("box/%s/version/%v/provider/%s", box.Tag, version.Version, provider.Name)

	providerCheck := &Provider{}

	ui.Say(fmt.Sprintf("Verifying provider upload: %s", provider.Name))

	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)

	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking token match for provider.. (attempt: %d)", attempts)

			resp, err := client.Get(path)

			if err != nil || (resp.StatusCode != 200) {
				cloudErrors := &VagrantCloudErrors{}
				err = decodeBody(resp, cloudErrors)
				err = fmt.Errorf("Error retrieving provider: %s", cloudErrors.FormatErrors())
				result <- err
				return
			}

			if err = decodeBody(resp, providerCheck); err != nil {
				err = fmt.Errorf("Error parsing provider response: %s", err)
				result <- err
				return
			}

			if err != nil {
				result <- err
				return
			}

			if upload.Token == providerCheck.HostedToken {
				// success!
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	ui.Message("Waiting for upload token match")
	log.Printf("Waiting for up to 600 seconds for provider hosted token to match %s", upload.Token)

	select {
	case err := <-result:
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("Upload succesfully verified with token %s", providerCheck.HostedToken))
		log.Printf("Box succesfully verified %s == %s", upload.Token, providerCheck.HostedToken)

		return multistep.ActionContinue
	case <-time.After(600 * time.Second):
		state.Put("error", fmt.Errorf("Timeout while waiting to for upload to verify token '%s'", upload.Token))
		return multistep.ActionHalt
	}
}

func (s *stepVerifyUpload) Cleanup(state multistep.StateBag) {
	// No cleanup
}

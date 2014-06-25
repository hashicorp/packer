package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepReleaseVersion struct {
}

func (s *stepReleaseVersion) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)
	config := state.Get("config").(Config)

	if config.NoRelease {
		ui.Say(fmt.Sprintf("Not releasing version due to configuration: %s", version.Version))
		return multistep.ActionContinue
	}

	path := fmt.Sprintf("box/%s/version/%v/release", box.Tag, version.Number)

	ui.Say(fmt.Sprintf("Releasing version: %s", version.Version))

	resp, err := client.Put(path)

	if err != nil || (resp.StatusCode != 200) {
		cloudErrors := &VagrantCloudErrors{}
		err = decodeBody(resp, cloudErrors)
		state.Put("error", fmt.Errorf("Error releasing version: %s", cloudErrors.FormatErrors()))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepReleaseVersion) Cleanup(state multistep.StateBag) {
	// No cleanup
}

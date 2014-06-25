package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type Version struct {
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Number      uint   `json:"number,omitempty"`
}

type stepCreateVersion struct {
	number uint // number of the version, if needed in cleanup
}

func (s *stepCreateVersion) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	box := state.Get("box").(*Box)

	ui.Say(fmt.Sprintf("Creating version: %s", config.Version))

	if hasVersion, v := box.HasVersion(config.Version); hasVersion {
		ui.Message(fmt.Sprintf("Version exists, skipping creation"))
		state.Put("version", v)
		return multistep.ActionContinue
	}

	path := fmt.Sprintf("box/%s/versions", box.Tag)

	version := &Version{Version: config.Version, Description: config.VersionDescription}

	// Wrap the version in a version object for the API
	wrapper := make(map[string]interface{})
	wrapper["version"] = version

	resp, err := client.Post(path, wrapper)

	if err != nil || (resp.StatusCode != 200) {
		cloudErrors := &VagrantCloudErrors{}
		err = decodeBody(resp, cloudErrors)
		state.Put("error", fmt.Errorf("Error creating version: %s", cloudErrors.FormatErrors()))
		return multistep.ActionHalt
	}

	if err = decodeBody(resp, version); err != nil {
		state.Put("error", fmt.Errorf("Error parsing version response: %s", err))
		return multistep.ActionHalt
	}

	// Save the number for cleanup
	s.number = version.Number

	state.Put("version", version)

	return multistep.ActionContinue
}

func (s *stepCreateVersion) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	box := state.Get("box").(*Box)

	// If we didn't save the version number, it likely doesn't exist or
	// already existed
	if s.number == 0 {
		ui.Message("Version was not created or previously existed, not deleting")
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	// Return if we didn't cancel or halt, and thus need
	// no cleanup
	if !cancelled && !halted {
		return
	}

	path := fmt.Sprintf("box/%s/version/%v", box.Tag, s.number)

	ui.Say("Cleaning up version")
	ui.Message(fmt.Sprintf("Deleting version: %s", config.Version))

	// No need for resp from the cleanup DELETE
	_, err := client.Delete(path)

	if err != nil {
		ui.Error(fmt.Sprintf("Error destroying version: %s", err))
	}

}

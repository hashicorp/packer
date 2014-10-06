package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type Provider struct {
	Name        string `json:"name"`
	HostedToken string `json:"hosted_token,omitempty"`
	UploadUrl   string `json:"upload_url,omitempty"`
}

type stepCreateProvider struct {
	name string // the name of the provider
}

func (s *stepCreateProvider) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)
	providerName := state.Get("providerName").(string)

	path := fmt.Sprintf("box/%s/version/%v/providers", box.Tag, version.Number)

	provider := &Provider{Name: providerName}

	// Wrap the provider in a provider object for the API
	wrapper := make(map[string]interface{})
	wrapper["provider"] = provider

	ui.Say(fmt.Sprintf("Creating provider: %s", providerName))

	resp, err := client.Post(path, wrapper)

	if err != nil || (resp.StatusCode != 200) {
		cloudErrors := &VagrantCloudErrors{}
		err = decodeBody(resp, cloudErrors)
		state.Put("error", fmt.Errorf("Error creating provider: %s", cloudErrors.FormatErrors()))
		return multistep.ActionHalt
	}

	if err = decodeBody(resp, provider); err != nil {
		state.Put("error", fmt.Errorf("Error parsing provider response: %s", err))
		return multistep.ActionHalt
	}

	// Save the name for cleanup
	s.name = provider.Name

	state.Put("provider", provider)

	return multistep.ActionContinue
}

func (s *stepCreateProvider) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	box := state.Get("box").(*Box)
	version := state.Get("version").(*Version)

	// If we didn't save the provider name, it likely doesn't exist
	if s.name == "" {
		ui.Say("Cleaning up provider")
		ui.Message("Provider was not created, not deleting")
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	// Return if we didn't cancel or halt, and thus need
	// no cleanup
	if !cancelled && !halted {
		return
	}

	ui.Say("Cleaning up provider")
	ui.Message(fmt.Sprintf("Deleting provider: %s", s.name))

	path := fmt.Sprintf("box/%s/version/%v/provider/%s", box.Tag, version.Number, s.name)

	// No need for resp from the cleanup DELETE
	_, err := client.Delete(path)

	if err != nil {
		ui.Error(fmt.Sprintf("Error destroying provider: %s", err))
	}
}

package vagrantcloud

import (
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type Box struct {
	Tag      string     `json:"tag"`
	Versions []*Version `json:"versions"`
}

func (b *Box) HasVersion(version string) (bool, *Version) {
	for _, v := range b.Versions {
		if v.Version == version {
			return true, v
		}
	}
	return false, nil
}

type stepVerifyBox struct {
}

func (s *stepVerifyBox) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VagrantCloudClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)

	ui.Say(fmt.Sprintf("Verifying box is accessible: %s", config.Tag))

	path := fmt.Sprintf("box/%s", config.Tag)
	resp, err := client.Get(path)

	if err != nil {
		state.Put("error", fmt.Errorf("Error retrieving box: %s", err))
		return multistep.ActionHalt
	}

	if resp.StatusCode != 200 {
		cloudErrors := &VagrantCloudErrors{}
		err = decodeBody(resp, cloudErrors)
		state.Put("error", fmt.Errorf("Error retrieving box: %s", cloudErrors.FormatErrors()))
		return multistep.ActionHalt
	}

	box := &Box{}

	if err = decodeBody(resp, box); err != nil {
		state.Put("error", fmt.Errorf("Error parsing box response: %s", err))
		return multistep.ActionHalt
	}

	if box.Tag != config.Tag {
		state.Put("error", fmt.Errorf("Could not verify box: %s", config.Tag))
		return multistep.ActionHalt
	}

	ui.Message("Box accessible and matches tag")

	// Keep the box in state for later
	state.Put("box", box)

	// Box exists and is accessible
	return multistep.ActionContinue
}

func (s *stepVerifyBox) Cleanup(state multistep.StateBag) {
	// no cleanup needed
}

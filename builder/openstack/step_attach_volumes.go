package openstack

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepAttachVolumes struct {
	VolumeAttachments []string
}

func (s *StepAttachVolumes) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	server := state.Get("server").(*servers.Server)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if len(s.VolumeAttachments) != 0 {
		ui.Say(fmt.Sprintf("Attaching volumes to server..."))
		for _, volumeId := range s.VolumeAttachments {
			_, err = volumeattach.Create(client, server.ID, volumeattach.CreateOpts{
				VolumeID: volumeId,
			}).Extract()
			if err != nil {
				err := fmt.Errorf(
					"Error attaching volume %s with instance: %s",
					volumeId, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Message(fmt.Sprintf("Attached volume %s to instance!", volumeId))
		}
	}

	return multistep.ActionContinue
}

func (s *StepAttachVolumes) Cleanup(state multistep.StateBag) {

}

package classic

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreatePVBuilder struct {
	Name              string
	BuilderVolumeName string
	SecurityListKey   string
}

func (s *stepCreatePVBuilder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating builder instance...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)
	ipAddName := state.Get("ipres_name").(string)
	secListName := state.Get(s.SecurityListKey).(string)

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:  s.Name,
		Shape: config.BuilderShape,
		Networking: map[string]compute.NetworkingInfo{
			"eth0": {
				Nat:      []string{ipAddName},
				SecLists: []string{secListName},
			},
		},
		Storage: []compute.StorageAttachmentInput{
			{
				Volume: s.BuilderVolumeName,
				Index:  1,
			},
		},
		ImageList: config.BuilderImageList,
		SSHKeys:   []string{config.Comm.SSHKeyPairName},
		Entry:     *config.BuilderImageListEntry,
	}

	instanceInfo, err := instanceClient.CreateInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Printf("Created instance %s", instanceInfo.Name)

	state.Put("builder_instance_info", instanceInfo)
	state.Put("builder_instance_id", instanceInfo.ID)

	ui.Message(fmt.Sprintf("Created builder instance: %s.", instanceInfo.Name))
	return multistep.ActionContinue
}

func (s *stepCreatePVBuilder) Cleanup(state multistep.StateBag) {
	instanceID, ok := state.GetOk("builder_instance_id")
	if !ok {
		return
	}

	// terminate instance
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*compute.Client)

	ui.Say("Terminating builder instance...")

	instanceClient := client.Instances()
	input := &compute.DeleteInstanceInput{
		Name: s.Name,
		ID:   instanceID.(string),
	}

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem destroying instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	ui.Say("Terminated builder instance.")

}

package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateTags struct {
	Tags map[string]string
}

func (s *StepCreateTags) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)

	if len(s.Tags) > 0 {
		for region, ami := range amis {
			ui.Say(fmt.Sprintf("Preparing tags for AMI (%s) and related snapshots", ami))

			// Declare list of resources to tag
			resourceIds := []string{ami}

			// Retrieve image list for given AMI
			imageResp, err := ec2conn.Images([]string{ami}, ec2.NewFilter())
			if err != nil {
				err := fmt.Errorf("Error retrieving details for AMI (%s): %s", ami, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			image := &imageResp.Images[0]

			// Add only those with a Snapshot ID, i.e. not Ephemeral
			for _, device := range image.BlockDevices {
				if device.SnapshotId != "" {
					ui.Say(fmt.Sprintf("Tagging snapshot: %s", device.SnapshotId))
					resourceIds = append(resourceIds, device.SnapshotId)
				}
			}

			var ec2Tags []*ec2.Tag
			for key, value := range s.Tags {
				ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"", key, value))
				ec2Tags = append(ec2Tags, &ec2.Tag{Key: &key, Value: &value})
			}

			_, err := regionconn.CreateTags(&ec2.CreateTagsInput{
				Resources: resourceIds,
				Tags:      ec2Tags,
			})
			if err != nil {
				err := fmt.Errorf("Error adding tags to AMI (%s): %s", ami, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

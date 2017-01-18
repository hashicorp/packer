package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

type stepTagEBSVolumes struct {
	VolumeRunTags map[string]string
}

func (s *stepTagEBSVolumes) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	if len(s.VolumeRunTags) > 0 {

		volumeIds := make([]*string, 0)
		for _, v := range instance.BlockDeviceMappings {
			if ebs := v.Ebs; ebs != nil {
				volumeIds = append(volumeIds, ebs.VolumeId)
			}
		}

		if len(volumeIds) == 0 {
			return multistep.ActionContinue
		}

		ui.Say(fmt.Sprintf("Adding tags to source EBS Volumes:"))
		tags := common.ConvertToEC2Tags(s.VolumeRunTags, ui)

		_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{
				instance.BlockDeviceMappings[0].Ebs.VolumeId,
			},
			Tags: tags,
		})
		if err != nil {
			err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepTagEBSVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

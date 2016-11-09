package ebsvolume

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepTagEBSVolumes struct {
	VolumeMapping []BlockDevice
}

func (s *stepTagEBSVolumes) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	volumes := make(EbsVolumes)
	for _, instanceBlockDevices := range instance.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			if configVolumeMapping.DeviceName == *instanceBlockDevices.DeviceName {
				volumes[*ec2conn.Config.Region] = append(
					volumes[*ec2conn.Config.Region],
					*instanceBlockDevices.Ebs.VolumeId)
			}
		}
	}
	state.Put("ebsvolumes", volumes)

	if len(s.VolumeMapping) > 0 {
		ui.Say("Tagging EBS volumes...")

		toTag := map[string][]*ec2.Tag{}
		for _, mapping := range s.VolumeMapping {
			if len(mapping.Tags) == 0 {
				ui.Say(fmt.Sprintf("No tags specified for volume on %s...", mapping.DeviceName))
				continue
			}

			tags := make([]*ec2.Tag, 0, len(mapping.Tags))
			for key, value := range mapping.Tags {
				tags = append(tags, &ec2.Tag{
					Key:   aws.String(fmt.Sprintf("%s", key)),
					Value: aws.String(fmt.Sprintf("%s", value)),
				})
			}

			for _, v := range instance.BlockDeviceMappings {
				if *v.DeviceName == mapping.DeviceName {
					toTag[*v.Ebs.VolumeId] = tags
				}
			}
		}

		for volumeId, tags := range toTag {
			_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
				Resources: []*string{&volumeId},
				Tags:      tags,
			})
			if err != nil {
				err := fmt.Errorf("Error tagging EBS Volume %s on %s: %s", volumeId, *instance.InstanceId, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		}
	}

	return multistep.ActionContinue
}

func (s *stepTagEBSVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

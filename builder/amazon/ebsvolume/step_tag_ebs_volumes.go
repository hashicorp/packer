package ebsvolume

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type stepTagEBSVolumes struct {
	VolumeMapping []BlockDevice
	Ctx           interpolate.Context
}

func (s *stepTagEBSVolumes) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	sourceAMI := state.Get("source_image").(*ec2.Image)
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
				s.Ctx.Data = &awscommon.BuildInfoTemplate{
					SourceAMI:   *sourceAMI.ImageId,
					BuildRegion: *ec2conn.Config.Region,
				}
				interpolatedValue, err := interpolate.Render(value, &s.Ctx)
				if err != nil {
					err = fmt.Errorf("Error processing tag: %s:%s - %s", key, value, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}

				tags = append(tags, &ec2.Tag{
					Key:   aws.String(fmt.Sprintf("%s", key)),
					Value: aws.String(fmt.Sprintf("%s", interpolatedValue)),
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

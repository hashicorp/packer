package ebsvolume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepTagEBSVolumes struct {
	VolumeMapping []BlockDevice
	Ctx           interpolate.Context
}

func (s *stepTagEBSVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	volumes := make(EbsVolumes)
	for _, instanceBlockDevices := range instance.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			if configVolumeMapping.DeviceName == *instanceBlockDevices.DeviceName {
				if configVolumeMapping.DeleteOnTermination {
					continue
				}
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

			ui.Message(fmt.Sprintf("Compiling list of tags to apply to volume on %s...", mapping.DeviceName))
			tags, err := mapping.Tags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
			if err != nil {
				err := fmt.Errorf("Error generating tags for device %s: %s", mapping.DeviceName, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			tags.Report(ui)

			// Generate the map of volumes and associated tags to apply.
			// Looping over the instance block device mappings allows us to
			// obtain the volumeId
			for _, v := range instance.BlockDeviceMappings {
				if *v.DeviceName == mapping.DeviceName {
					toTag[*v.Ebs.VolumeId] = tags
				}
			}
		}

		// Apply the tags
		for volumeId, tags := range toTag {
			ui.Message(fmt.Sprintf("Applying tags to EBS Volume: %s", volumeId))
			_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
				Resources: aws.StringSlice([]string{volumeId}),
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

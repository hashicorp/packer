package ebsvolume

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer-plugin-amazon/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type stepTagEBSVolumes struct {
	VolumeMapping []BlockDevice
	Ctx           interpolate.Context
}

func (s *stepTagEBSVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

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
		// If run_volume_tags were set in the template any attached EBS
		// volume will have had these tags applied when the instance was
		// created. We now need to remove these tags to ensure only the EBS
		// volume tags are applied (if any)
		if len(config.VolumeRunTags) > 0 {
			ui.Say("Removing any tags applied to EBS volumes when the source instance was created...")

			ui.Message("Compiling list of existing tags to remove...")
			existingTags, err := awscommon.TagMap(config.VolumeRunTags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
			if err != nil {
				err := fmt.Errorf("Error generating list of tags to remove: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			existingTags.Report(ui)

			// Generate the list of volumes with tags to delete.
			// Looping over the instance block device mappings allows us to
			// obtain the volumeId
			volumeIds := []string{}
			for _, mapping := range s.VolumeMapping {
				for _, v := range instance.BlockDeviceMappings {
					if *v.DeviceName == mapping.DeviceName {
						volumeIds = append(volumeIds, *v.Ebs.VolumeId)
					}
				}
			}

			// Delete the tags
			ui.Message(fmt.Sprintf("Deleting 'run_volume_tags' on EBS Volumes: %s", strings.Join(volumeIds, ", ")))
			_, err = ec2conn.DeleteTags(&ec2.DeleteTagsInput{
				Resources: aws.StringSlice(volumeIds),
				Tags:      existingTags,
			})
			if err != nil {
				err := fmt.Errorf("Error deleting tags on EBS Volumes %s: %s", strings.Join(volumeIds, ", "), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		ui.Say("Tagging EBS volumes...")
		toTag := map[string][]*ec2.Tag{}
		for _, mapping := range s.VolumeMapping {
			if len(mapping.Tags) == 0 {
				ui.Say(fmt.Sprintf("No tags specified for volume on %s...", mapping.DeviceName))
				continue
			}

			ui.Message(fmt.Sprintf("Compiling list of tags to apply to volume on %s...", mapping.DeviceName))
			tags, err := awscommon.TagMap(mapping.Tags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
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

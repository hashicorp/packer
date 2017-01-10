package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type stepTagEBSVolumes struct {
	VolumeRunTags map[string]string
	Ctx           interpolate.Context
}

func (s *stepTagEBSVolumes) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	sourceAMI := state.Get("source_image").(*ec2.Image)
	ui := state.Get("ui").(packer.Ui)

	if len(s.VolumeRunTags) == 0 {
		return multistep.ActionContinue
	}

	volumeIds := make([]*string, 0)
	for _, v := range instance.BlockDeviceMappings {
		if ebs := v.Ebs; ebs != nil {
			volumeIds = append(volumeIds, ebs.VolumeId)
		}
	}

	if len(volumeIds) == 0 {
		return multistep.ActionContinue
	}

	tags := make([]*ec2.Tag, len(s.VolumeRunTags))
	for key, value := range s.VolumeRunTags {
		s.Ctx.Data = &common.BuildInfoTemplate{
			SourceAMI:   *sourceAMI.ImageId,
			BuildRegion: *ec2conn.Config.Region,
		}
		interpolatedValue, err := interpolate.Render(value, &s.Ctx)
		if err != nil {
			err = fmt.Errorf("Error processing volume tag: %s:%s - %s", key, value, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		tags = append(tags, &ec2.Tag{Key: &key, Value: &interpolatedValue})
	}

	ui.Say("Adding tags to source EBS Volumes")
	tags, err := common.ConvertToEC2Tags(s.VolumeRunTags, *ec2conn.Config.Region, *sourceAMI.ImageId, s.Ctx, ui)
	if err != nil {
		err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
		Resources: volumeIds,
		Tags:      tags,
	})
	if err != nil {
		err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepTagEBSVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// stepCleanupVolumes cleans up any orphaned volumes that were not designated to
// remain after termination of the instance. These volumes are typically ones
// that are marked as "delete on terminate:false" in the source_ami of a build.
type StepCleanupVolumes struct {
	LaunchMappings BlockDevices
}

func (s *StepCleanupVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// stepCleanupVolumes is for Cleanup only
	return multistep.ActionContinue
}

func (s *StepCleanupVolumes) Cleanup(state multistep.StateBag) {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instanceRaw := state.Get("instance")
	var instance *ec2.Instance
	if instanceRaw != nil {
		instance = instanceRaw.(*ec2.Instance)
	}
	ui := state.Get("ui").(packersdk.Ui)
	if instance == nil {
		ui.Say("No volumes to clean up, skipping")
		return
	}

	ui.Say("Cleaning up any extra volumes...")

	// Collect Volume information from the cached Instance as a map of volume-id
	// to device name, to compare with save list below
	var vl []*string
	volList := make(map[string]string)
	for _, bdm := range instance.BlockDeviceMappings {
		if bdm.Ebs != nil {
			vl = append(vl, bdm.Ebs.VolumeId)
			volList[*bdm.Ebs.VolumeId] = *bdm.DeviceName
		}
	}

	// Using the volume list from the cached Instance, check with AWS for up to
	// date information on them
	resp, err := ec2conn.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("volume-id"),
				Values: vl,
			},
		},
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error describing volumes: %s", err))
		return
	}

	// If any of the returned volumes are in a "deleting" stage or otherwise not
	// available, remove them from the list of volumes
	for _, v := range resp.Volumes {
		if v.State != nil && *v.State != "available" {
			delete(volList, *v.VolumeId)
		}
	}

	if len(resp.Volumes) == 0 {
		ui.Say("No volumes to clean up, skipping")
		return
	}

	// Filter out any devices created as part of the launch mappings, since
	// we'll let amazon follow the `delete_on_termination` setting.
	for _, b := range s.LaunchMappings {
		for volKey, volName := range volList {
			if volName == b.DeviceName {
				delete(volList, volKey)
			}
		}
	}

	// Destroy remaining volumes
	for k := range volList {
		ui.Say(fmt.Sprintf("Destroying volume (%s)...", k))
		_, err := ec2conn.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: aws.String(k)})
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting volume: %s", err))
		}

	}
}

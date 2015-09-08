package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// stepCleanupVolumes cleans up any orphaned volumes that were not designated to
// remain after termination of the instance. These volumes are typically ones
// that are marked as "delete on terminate:false" in the source_ami of a build.
type stepCleanupVolumes struct {
	BlockDevices common.BlockDevices
}

func (s *stepCleanupVolumes) Run(state multistep.StateBag) multistep.StepAction {
	// stepCleanupVolumes is for Cleanup only
	return multistep.ActionContinue
}

func (s *stepCleanupVolumes) Cleanup(state multistep.StateBag) {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instanceRaw := state.Get("instance")
	var instance *ec2.Instance
	if instanceRaw != nil {
		instance = instanceRaw.(*ec2.Instance)
	}
	ui := state.Get("ui").(packer.Ui)
	amisRaw := state.Get("amis")
	if amisRaw == nil {
		ui.Say("No AMIs to cleanup")
		return
	}

	if instance == nil {
		ui.Say("No volumes to clean up, skipping")
		return
	}

	ui.Say("Cleaning up any extra volumes...")

	// We don't actually care about the value here, but we need Set behavior
	save := make(map[string]struct{})
	for _, b := range s.BlockDevices.AMIMappings {
		if !b.DeleteOnTermination {
			save[b.DeviceName] = struct{}{}
		}
	}

	for _, b := range s.BlockDevices.LaunchMappings {
		if !b.DeleteOnTermination {
			save[b.DeviceName] = struct{}{}
		}
	}

	// Collect Volume information from the cached Instance as a map of volume-id
	// to device name, to compare with save list above
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
			&ec2.Filter{
				Name:   aws.String("volume-id"),
				Values: vl,
			},
		},
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Error describing volumes: %s", err))
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

	// Filter out any devices marked for saving
	for saveName, _ := range save {
		for volKey, volName := range volList {
			if volName == saveName {
				delete(volList, volKey)
			}
		}
	}

	// Destroy remaining volumes
	for k, _ := range volList {
		ui.Say(fmt.Sprintf("Destroying volume (%s)...", k))
		_, err := ec2conn.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: aws.String(k)})
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting volume: %s", k))
		}

	}
}

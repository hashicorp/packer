package ebsvolume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepSnapshotEBSVolumes struct {
	VolumeMapping []BlockDevice
	Ctx           interpolate.Context
}

func (s *stepSnapshotEBSVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)
	snapshotsIds := make([]string, 0)

	for _, instanceBlockDevices := range instance.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			//Find the config entry for the instance blockDevice
			if configVolumeMapping.DeviceName == *instanceBlockDevices.DeviceName {
				if configVolumeMapping.SnapshotVolume != true {
					continue
				}

				ui.Message(fmt.Sprintf("Compiling list of tags to apply to snapshot from Volume %s...", *instanceBlockDevices.DeviceName))
				tags, err := awscommon.TagMap(configVolumeMapping.Tags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
				if err != nil {
					err := fmt.Errorf("Error generating tags for device %s: %s", *instanceBlockDevices.DeviceName, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}

				tagSpec := &ec2.TagSpecification{
					ResourceType: aws.String("snapshot"),
					Tags:         tags,
				}

				input := &ec2.CreateSnapshotInput{
					VolumeId:          instanceBlockDevices.Ebs.VolumeId,
					TagSpecifications: []*ec2.TagSpecification{tagSpec},
				}
				snapshot, err := ec2conn.CreateSnapshot(input)
				if err != nil {
					err := fmt.Errorf("Error generating snapsot for volume %s: %s", *instanceBlockDevices.Ebs.VolumeId, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				snapshotsIds = append(snapshotsIds, *snapshot.SnapshotId)
			}
		}
	}

	ui.Say("Waiting for Snapshots to become ready...")
	for _, snapID := range snapshotsIds {
		ui.Message(fmt.Sprintf("Waiting for %s to be ready.", snapID))
		err := awscommon.WaitUntilSnapshotDone(ctx, ec2conn, snapID)
		if err != nil {
			err = fmt.Errorf("Error waiting for snapsot to become ready %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			ui.Message("Failed to wait")
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Snapshot Ready: %s", snapID))
	}

	return multistep.ActionContinue
}

func (s *stepSnapshotEBSVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

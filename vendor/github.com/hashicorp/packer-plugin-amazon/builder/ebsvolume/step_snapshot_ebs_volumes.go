package ebsvolume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	awscommon "github.com/hashicorp/packer-plugin-amazon/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type stepSnapshotEBSVolumes struct {
	PollingConfig *awscommon.AWSPollingConfig
	AccessConfig  *awscommon.AccessConfig
	VolumeMapping []BlockDevice
	//Map of SnapshotID: BlockDevice, Where *BlockDevice is in VolumeMapping
	snapshotMap map[string]*BlockDevice
	Ctx         interpolate.Context
}

func (s *stepSnapshotEBSVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(ec2iface.EC2API)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	s.snapshotMap = make(map[string]*BlockDevice)

	for _, instanceBlockDevice := range instance.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			//Find the config entry for the instance blockDevice
			if configVolumeMapping.DeviceName == *instanceBlockDevice.DeviceName {
				//Skip Volumes that are not set to create snapshot
				if configVolumeMapping.SnapshotVolume != true {
					continue
				}

				ui.Message(fmt.Sprintf("Compiling list of tags to apply to snapshot from Volume %s...", *instanceBlockDevice.DeviceName))
				tags, err := awscommon.TagMap(configVolumeMapping.SnapshotTags).EC2Tags(s.Ctx, s.AccessConfig.SessionRegion(), state)
				if err != nil {
					err := fmt.Errorf("Error generating tags for snapshot %s: %s", *instanceBlockDevice.DeviceName, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				tags.Report(ui)

				tagSpec := &ec2.TagSpecification{
					ResourceType: aws.String("snapshot"),
					Tags:         tags,
				}

				input := &ec2.CreateSnapshotInput{
					VolumeId:          aws.String(*instanceBlockDevice.Ebs.VolumeId),
					TagSpecifications: []*ec2.TagSpecification{tagSpec},
				}

				//Dont try to set an empty tag spec
				if len(tags) == 0 {
					input.TagSpecifications = nil
				}

				ui.Message(fmt.Sprintf("Requesting snapshot of volume: %s...", *instanceBlockDevice.Ebs.VolumeId))
				snapshot, err := ec2conn.CreateSnapshot(input)
				if err != nil || snapshot == nil {
					err := fmt.Errorf("Error generating snapsot for volume %s: %s", *instanceBlockDevice.Ebs.VolumeId, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				ui.Message(fmt.Sprintf("Requested Snapshot of Volume %s: %s", *instanceBlockDevice.Ebs.VolumeId, *snapshot.SnapshotId))
				s.snapshotMap[*snapshot.SnapshotId] = &configVolumeMapping
			}
		}
	}

	ui.Say("Waiting for Snapshots to become ready...")
	for snapID := range s.snapshotMap {
		ui.Message(fmt.Sprintf("Waiting for %s to be ready.", snapID))
		err := s.PollingConfig.WaitUntilSnapshotDone(ctx, ec2conn, snapID)
		if err != nil {
			err = fmt.Errorf("Error waiting for snapsot to become ready %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			ui.Message("Failed to wait")
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Snapshot Ready: %s", snapID))
	}

	//Attach User and Group permissions to snapshots
	ui.Say("Setting User/Group Permissions for Snapshots...")
	for snapID, bd := range s.snapshotMap {
		snapshotOptions := make(map[string]*ec2.ModifySnapshotAttributeInput)

		if len(bd.SnapshotGroups) > 0 {
			groups := make([]*string, len(bd.SnapshotGroups))
			addsSnapshot := make([]*ec2.CreateVolumePermission, len(bd.SnapshotGroups))

			addSnapshotGroups := &ec2.ModifySnapshotAttributeInput{
				CreateVolumePermission: &ec2.CreateVolumePermissionModifications{},
			}

			for i, g := range bd.SnapshotGroups {
				groups[i] = aws.String(g)
				addsSnapshot[i] = &ec2.CreateVolumePermission{
					Group: aws.String(g),
				}
			}

			addSnapshotGroups.GroupNames = groups
			addSnapshotGroups.CreateVolumePermission.Add = addsSnapshot
			snapshotOptions["groups"] = addSnapshotGroups

		}

		if len(bd.SnapshotUsers) > 0 {
			users := make([]*string, len(bd.SnapshotUsers))
			addsSnapshot := make([]*ec2.CreateVolumePermission, len(bd.SnapshotUsers))
			for i, u := range bd.SnapshotUsers {
				users[i] = aws.String(u)
				addsSnapshot[i] = &ec2.CreateVolumePermission{UserId: aws.String(u)}
			}

			snapshotOptions["users"] = &ec2.ModifySnapshotAttributeInput{
				UserIds: users,
				CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
					Add: addsSnapshot,
				},
			}
		}

		//Todo: Copy to other regions and repeat this block in all regions.
		for name, input := range snapshotOptions {
			ui.Message(fmt.Sprintf("Modifying: %s", name))
			input.SnapshotId = &snapID
			_, err := ec2conn.ModifySnapshotAttribute(input)
			if err != nil {
				err := fmt.Errorf("Error modify snapshot attributes: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	//Record all snapshots in current Region.
	snapshots := make(EbsSnapshots)
	currentregion := s.AccessConfig.SessionRegion()

	for snapID := range s.snapshotMap {
		snapshots[currentregion] = append(
			snapshots[currentregion],
			snapID)
	}
	//Records artifacts
	state.Put("ebssnapshots", snapshots)

	return multistep.ActionContinue
}

func (s *stepSnapshotEBSVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

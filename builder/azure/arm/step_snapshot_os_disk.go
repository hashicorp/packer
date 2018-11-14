package arm

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSnapshotOSDisk struct {
	client         *AzureClient
	create         func(ctx context.Context, resourceGroupName string, srcUriVhd string, location string, tags map[string]*string, dstSnapshotName string) error
	say            func(message string)
	error          func(e error)
	isManagedImage bool
}

func NewStepSnapshotOSDisk(client *AzureClient, ui packer.Ui, isManagedImage bool) *StepSnapshotOSDisk {
	var step = &StepSnapshotOSDisk{
		client:         client,
		say:            func(message string) { ui.Say(message) },
		error:          func(e error) { ui.Error(e.Error()) },
		isManagedImage: isManagedImage,
	}

	step.create = step.createSnapshot
	return step
}

func (s *StepSnapshotOSDisk) createSnapshot(ctx context.Context, resourceGroupName string, srcUriVhd string, location string, tags map[string]*string, dstSnapshotName string) error {

	srcVhdToSnapshot := compute.Snapshot{
		DiskProperties: &compute.DiskProperties{
			CreationData: &compute.CreationData{
				CreateOption:     compute.Copy,
				SourceResourceID: to.StringPtr(srcUriVhd),
			},
		},
		Location: to.StringPtr(location),
		Tags:     tags,
	}

	f, err := s.client.SnapshotsClient.CreateOrUpdate(ctx, resourceGroupName, dstSnapshotName, srcVhdToSnapshot)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	err = f.WaitForCompletion(ctx, s.client.SnapshotsClient.Client)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	createdSnapshot, err := f.Result(s.client.SnapshotsClient)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	s.say(fmt.Sprintf(" -> Managed Image OS Disk Snapshot		: '%s'", *(createdSnapshot.ID)))

	return nil
}

func (s *StepSnapshotOSDisk) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	if s.isManagedImage {

		s.say("Taking snapshot of OS disk ...")

		var resourceGroupName = stateBag.Get(constants.ArmManagedImageResourceGroupName).(string)
		var location = stateBag.Get(constants.ArmLocation).(string)
		var tags = stateBag.Get(constants.ArmTags).(map[string]*string)
		var srcUriVhd = stateBag.Get(constants.ArmOSDiskVhd).(string)
		var dstSnapshotName = stateBag.Get(constants.ArmManagedImageOSDiskSnapshotName).(string)

		err := s.create(ctx, resourceGroupName, srcUriVhd, location, tags, dstSnapshotName)

		if err != nil {
			stateBag.Put(constants.Error, err)
			s.error(err)

			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (*StepSnapshotOSDisk) Cleanup(multistep.StateBag) {
}

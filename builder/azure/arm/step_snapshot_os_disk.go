package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSnapshotOSDisk struct {
	client *AzureClient
	create func(ctx context.Context, resourceGroupName string, srcUriVhd string, location string, tags map[string]*string, dstSnapshotName string) error
	say    func(message string)
	error  func(e error)
	enable func() bool
}

func NewStepSnapshotOSDisk(client *AzureClient, ui packersdk.Ui, config *Config) *StepSnapshotOSDisk {
	var step = &StepSnapshotOSDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
		enable: func() bool { return config.isManagedImage() && config.ManagedImageOSDiskSnapshotName != "" },
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

	err = f.WaitForCompletionRef(ctx, s.client.SnapshotsClient.Client)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	createdSnapshot, err := f.Result(s.client.SnapshotsClient)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	s.say(fmt.Sprintf(" -> Snapshot ID : '%s'", *(createdSnapshot.ID)))
	return nil
}

func (s *StepSnapshotOSDisk) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	if !s.enable() {
		return multistep.ActionContinue
	}

	s.say("Snapshotting OS disk ...")

	var resourceGroupName = stateBag.Get(constants.ArmManagedImageResourceGroupName).(string)
	var location = stateBag.Get(constants.ArmLocation).(string)
	var tags = stateBag.Get(constants.ArmTags).(map[string]*string)
	var srcUriVhd = stateBag.Get(constants.ArmOSDiskVhd).(string)
	var dstSnapshotName = stateBag.Get(constants.ArmManagedImageOSDiskSnapshotName).(string)

	s.say(fmt.Sprintf(" -> OS Disk     : '%s'", srcUriVhd))
	err := s.create(ctx, resourceGroupName, srcUriVhd, location, tags, dstSnapshotName)

	if err != nil {
		stateBag.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepSnapshotOSDisk) Cleanup(multistep.StateBag) {
}

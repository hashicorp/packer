package arm

import (
	"context"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSnapshotDataDisks struct {
	client *AzureClient
	create func(ctx context.Context, resourceGroupName string, srcUriVhd string, location string, tags map[string]*string, dstSnapshotName string) error
	say    func(message string)
	error  func(e error)
}

func NewStepSnapshotDataDisks(client *AzureClient, ui packer.Ui) *StepSnapshotDataDisks {
	var step = &StepSnapshotDataDisks{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.create = step.createDataDiskSnapshot
	return step
}

func (s *StepSnapshotDataDisks) createDataDiskSnapshot(ctx context.Context, resourceGroupName string, srcUriVhd string, location string, tags map[string]*string, dstSnapshotName string) error {

	srcVhdToSnapshot := compute.Snapshot{
		DiskProperties: &compute.DiskProperties{
			CreationData: &compute.CreationData{
				CreateOption: compute.Import,
				SourceURI:    to.StringPtr(srcUriVhd),
			},
		},
		Location: to.StringPtr(location),
		Tags:     tags,
	}

	f, err := s.client.SnapshotsClient.CreateOrUpdate(ctx, resourceGroupName, dstSnapshotName, srcVhdToSnapshot)

	if err != nil {
		s.say(s.client.LastError.Error())
	}

	return f.WaitForCompletion(ctx, s.client.SnapshotsClient.Client)
}

func (s *StepSnapshotDataDisks) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	s.say("Taking snapshot of OS disk ...")

	var resourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var location = stateBag.Get(constants.ArmLocation).(string)
	var tags = stateBag.Get(constants.ArmTags).(map[string]*string)
	var additionalDisks = stateBag.Get(constants.ArmAdditionalDiskVhds).([]string)
	var dstSnapshotPrefix = stateBag.Get(constants.ArmManagedImageDataDiskSnapshotPrefix).(string)

	for i, disk := range additionalDisks {
		dstSnapshotName := dstSnapshotPrefix + strconv.Itoa(i)
		err := s.create(ctx, resourceGroupName, disk, location, tags, dstSnapshotName)

		if err != nil {
			stateBag.Put(constants.Error, err)
			s.error(err)

			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (*StepSnapshotDataDisks) Cleanup(multistep.StateBag) {
}

package arm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/packer/builder/azure/common/constants"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDeleteAdditionalDisk struct {
	client        *AzureClient
	delete        func(string, string) error
	deleteManaged func(context.Context, string, string) error
	say           func(message string)
	error         func(e error)
}

func NewStepDeleteAdditionalDisks(client *AzureClient, ui packersdk.Ui) *StepDeleteAdditionalDisk {
	var step = &StepDeleteAdditionalDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteBlob
	step.deleteManaged = step.deleteManagedDisk
	return step
}

func (s *StepDeleteAdditionalDisk) deleteBlob(storageContainerName string, blobName string) error {
	blob := s.client.BlobStorageClient.GetContainerReference(storageContainerName).GetBlobReference(blobName)
	_, err := blob.BreakLease(nil)
	if err != nil && !strings.Contains(err.Error(), "LeaseNotPresentWithLeaseOperation") {
		s.say(s.client.LastError.Error())
		return err
	}

	err = blob.Delete(nil)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepDeleteAdditionalDisk) deleteManagedDisk(ctx context.Context, resourceGroupName string, diskName string) error {
	xs := strings.Split(diskName, "/")
	diskName = xs[len(xs)-1]
	f, err := s.client.DisksClient.Delete(ctx, resourceGroupName, diskName)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DisksClient.Client)
	}
	return err
}

func (s *StepDeleteAdditionalDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deleting the temporary Additional disk ...")

	var dataDisks []string

	if disks := state.Get(constants.ArmAdditionalDiskVhds); disks != nil {
		dataDisks = disks.([]string)
	}
	var isManagedDisk = state.Get(constants.ArmIsManagedImage).(bool)
	var isExistingResourceGroup = state.Get(constants.ArmIsExistingResourceGroup).(bool)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	if dataDisks == nil {
		s.say(fmt.Sprintf(" -> No Additional Disks specified"))
		return multistep.ActionContinue
	}

	if isManagedDisk && !isExistingResourceGroup {
		s.say(fmt.Sprintf(" -> Additional Disk : skipping, managed disk was used..."))
		return multistep.ActionContinue
	}

	for i, additionaldisk := range dataDisks {
		s.say(fmt.Sprintf(" -> Additional Disk %d: '%s'", i+1, additionaldisk))
		var err error
		if isManagedDisk {
			err = s.deleteManaged(ctx, resourceGroupName, additionaldisk)
			if err != nil {
				s.say("Failed to delete the managed Additional Disk!")
				return processStepResult(err, s.error, state)
			}
			continue
		}

		u, err := url.Parse(additionaldisk)
		if err != nil {
			s.say("Failed to parse the Additional Disk's VHD URI!")
			return processStepResult(err, s.error, state)
		}

		xs := strings.Split(u.Path, "/")
		if len(xs) < 3 {
			err = errors.New("Failed to parse Additional Disk's VHD URI!")
		} else {
			var storageAccountName = xs[1]
			var blobName = strings.Join(xs[2:], "/")

			err = s.delete(storageAccountName, blobName)
		}
		if err != nil {
			return processStepResult(err, s.error, state)
		}
	}
	return multistep.ActionContinue
}

func (*StepDeleteAdditionalDisk) Cleanup(multistep.StateBag) {
}

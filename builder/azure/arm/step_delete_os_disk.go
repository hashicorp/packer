package arm

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/packer/builder/azure/common/constants"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepDeleteOSDisk struct {
	client        *AzureClient
	delete        func(string, string) error
	deleteManaged func(string, string) error
	say           func(message string)
	error         func(e error)
}

func NewStepDeleteOSDisk(client *AzureClient, ui packer.Ui) *StepDeleteOSDisk {
	var step = &StepDeleteOSDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteBlob
	step.deleteManaged = step.deleteManagedDisk
	return step
}

func (s *StepDeleteOSDisk) deleteBlob(storageContainerName string, blobName string) error {
	blob := s.client.BlobStorageClient.GetContainerReference(storageContainerName).GetBlobReference(blobName)
	err := blob.Delete(nil)

	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepDeleteOSDisk) deleteManagedDisk(resourceGroupName string, imageName string) error {
	xs := strings.Split(imageName, "/")
	diskName := xs[len(xs)-1]
	_, errChan := s.client.DisksClient.Delete(resourceGroupName, diskName, nil)
	err := <-errChan
	return err
}

func (s *StepDeleteOSDisk) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting the temporary OS disk ...")

	var osDisk = state.Get(constants.ArmOSDiskVhd).(string)
	var isManagedDisk = state.Get(constants.ArmIsManagedImage).(bool)
	var isExistingResourceGroup = state.Get(constants.ArmIsExistingResourceGroup).(bool)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	if isManagedDisk && !isExistingResourceGroup {
		s.say(fmt.Sprintf(" -> OS Disk : skipping, managed disk was used..."))
		return multistep.ActionContinue
	}

	s.say(fmt.Sprintf(" -> OS Disk : '%s'", osDisk))

	var err error
	if isManagedDisk {
		err = s.deleteManaged(resourceGroupName, osDisk)
		if err != nil {
			s.say("Failed to delete the managed OS Disk!")
			return processStepResult(err, s.error, state)
		}
		return multistep.ActionContinue
	}
	u, err := url.Parse(osDisk)
	if err != nil {
		s.say("Failed to parse the OS Disk's VHD URI!")
		return processStepResult(err, s.error, state)
	}

	xs := strings.Split(u.Path, "/")
	if len(xs) < 3 {
		err = errors.New("Failed to parse OS Disk's VHD URI!")
	} else {
		var storageAccountName = xs[1]
		var blobName = strings.Join(xs[2:], "/")

		err = s.delete(storageAccountName, blobName)
	}
	return processStepResult(err, s.error, state)
}

func (*StepDeleteOSDisk) Cleanup(multistep.StateBag) {
}

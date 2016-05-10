// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mitchellh/packer/builder/azure/common/constants"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDeleteOSDisk struct {
	client *AzureClient
	delete func(string, string) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeleteOSDisk(client *AzureClient, ui packer.Ui) *StepDeleteOSDisk {
	var step = &StepDeleteOSDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteBlob
	return step
}

func (s *StepDeleteOSDisk) deleteBlob(storageContainerName string, blobName string) error {
	return s.client.BlobStorageClient.DeleteBlob(storageContainerName, blobName, nil)
}

func (s *StepDeleteOSDisk) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting the temporary OS disk ...")

	var osDisk = state.Get(constants.ArmOSDiskVhd).(string)
	s.say(fmt.Sprintf(" -> OS Disk : '%s'", osDisk))

	u, err := url.Parse(osDisk)
	if err != nil {
		s.say("Failed to parse the OS Disk's VHD URI!")
		return multistep.ActionHalt
	}

	xs := strings.Split(u.Path, "/")

	var storageAccountName = xs[1]
	var blobName = strings.Join(xs[2:], "/")

	err = s.delete(storageAccountName, blobName)
	return processStepResult(err, s.error, state)
}

func (*StepDeleteOSDisk) Cleanup(multistep.StateBag) {
}

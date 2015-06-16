// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
)

type StepCreateService struct {
	Location string
	TmpServiceName string
}

func (s *StepCreateService) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error creating temporary Azure service: %s"

	ui.Say("Creating temporary Azure service...")

	requestData := reqManager.CreateCloudService(s.TmpServiceName, s.Location)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.SrvExists, 1)

	return multistep.ActionContinue
}

func (s *StepCreateService) Cleanup(state multistep.StateBag) {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	var err error
	var res int

	if res = state.Get(constants.SrvExists).(int); res == 1 {
		ui.Say("Removing temporary Azure service and it's deployments if any...")
		errorMsg := "Error removing temporary Azure service: %s"

		var requestData *request.Data
		requestData = reqManager.DeleteCloudServiceAndMedia(s.TmpServiceName)

		err = reqManager.ExecuteSync(requestData)

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}
}

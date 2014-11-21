// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
)

type StepStopVm struct {
	TmpVmName string
	TmpServiceName string
}

func (s *StepStopVm) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error stopping temporary Azure VM: %s"

	ui.Say("Stopping temporary Azure VM...")

	requestData := reqManager.ShutdownRoles(s.TmpServiceName, s.TmpVmName)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmRunning, 0)

	return multistep.ActionContinue
}

func (s *StepStopVm) Cleanup(state multistep.StateBag) {
	// do nothing
}

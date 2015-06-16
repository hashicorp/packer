// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
)

type StepCreateVm struct {
	OsType string
	StorageAccount string
	StorageAccountContainer string
	OsImageLabel string
	TmpVmName string
	TmpServiceName string
	InstanceSize string
	Username string
	Password string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error creating temporary Azure VM: %s"
	var err error

	ui.Say("Creating temporary Azure VM...")

	osImageName := state.Get(constants.OsImageName).(string)
	if len(osImageName) == 0 {
		err := fmt.Errorf(errorMsg, fmt.Errorf("osImageName is empty"))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	isOSImage := state.Get(constants.IsOSImage).(bool)

	mediaLoc := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s.vhd", s.StorageAccount, s.StorageAccountContainer, s.TmpVmName)

	requestData := reqManager.CreateVirtualMachineDeploymentWin(isOSImage, s.TmpServiceName, s.TmpVmName, s.InstanceSize, s.Username, s.Password, osImageName, mediaLoc )

	err = reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmExists, 1)
	state.Put(constants.DiskExists, 1)

	return multistep.ActionContinue
}

func (s *StepCreateVm) Cleanup(state multistep.StateBag) {
	// do nothing
}

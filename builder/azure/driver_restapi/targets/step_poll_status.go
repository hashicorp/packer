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
	"log"
	"time"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response/model"
)

const(
	powerState_Started string = "Started"
	powerState_Stopping string = "Stopping"
	powerState_Stopped string = "Stopped"
	powerState_Unknown string = "Unknown"
	instanceStatus_ReadyRole = "ReadyRole"
	instanceStatus_FailedStartingRole = "FailedStartingRole"
	instanceStatus_FailedStartingVM = "FailedStartingVM"
	instanceStatus_ProvisioningFailed = "ProvisioningFailed"
	instanceStatus_UnresponsiveRole = "UnresponsiveRole"
)

type StepPollStatus struct {
	TmpServiceName string
	TmpVmName string
	OsType string
}

func (s *StepPollStatus) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error polling temporary Azure VM is ready: %s"

	ui.Say("Polling temporary Azure VM is ready...")

	if len(s.OsType) == 0 {
		err := fmt.Errorf(errorMsg, "'OsType' param is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	firstSleepMin := time.Duration(2)
	firstSleepTime := time.Minute * firstSleepMin
	log.Printf("Sleeping for %v min to make the VM to start", uint(firstSleepMin))
	time.Sleep(firstSleepTime)

	var count uint = 60
	var duration time.Duration = 40
	sleepTime := time.Second * duration
	total := count*uint(duration)

	//	var err error
	var deployment *model.Deployment

	requestData := reqManager.GetDeployment(s.TmpServiceName, s.TmpVmName)

	for count != 0 {
		resp, err := reqManager.Execute(requestData)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		deployment, err = response.ParseDeployment(resp.Body)
//		log.Printf("deployment:\n%v", deployment)

		if len(deployment.RoleInstanceList) > 0 {
			powerState := deployment.RoleInstanceList[0].PowerState
			instanceStatus := deployment.RoleInstanceList[0].InstanceStatus

			if powerState == powerState_Started && instanceStatus == instanceStatus_ReadyRole {
				break;
			}

			if instanceStatus == instanceStatus_FailedStartingRole ||
				instanceStatus == instanceStatus_FailedStartingVM ||
				instanceStatus == instanceStatus_ProvisioningFailed ||
				instanceStatus == instanceStatus_UnresponsiveRole {
				err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].instanceStatus is " + instanceStatus)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			if powerState == powerState_Stopping ||
				powerState == powerState_Stopped ||
				powerState == powerState_Unknown {
				err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].PowerState is " + powerState)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		// powerState_Starting or deployment.RoleInstanceList[0] == 0
		log.Println(fmt.Sprintf("Waiting for another %v seconds...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if(count == 0){
		err := fmt.Errorf(errorMsg, fmt.Sprintf("time is up (%d seconds)", total))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmRunning, 1)

	log.Println("s.OsType = " + s.OsType)

	if s.OsType == Linux {
		endpoints := deployment.RoleInstanceList[0].InstanceEndpoints
		if len(endpoints) == 0{
			err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].InstanceEndpoints list is empty")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		vip := endpoints[0].Vip
		port := endpoints[0].PublicPort
		endpoint := fmt.Sprintf("%s:%s", vip, port)

		ui.Message("VM endpoint: " + endpoint)
		state.Put(constants.AzureVmAddr, endpoint)
	}

	roleList := deployment.RoleList
	if len(roleList) == 0{
		err := fmt.Errorf(errorMsg, "deployment.RoleList is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	diskName := roleList[0].OSVirtualHardDisk.DiskName
	ui.Message("VM DiskName: " + diskName)
	state.Put(constants.HardDiskName, diskName)

	mediaLink := roleList[0].OSVirtualHardDisk.MediaLink
	ui.Message("VM MediaLink: " + mediaLink)
	state.Put(constants.MediaLink, mediaLink)

	return multistep.ActionContinue
}

func (s *StepPollStatus) Cleanup(state multistep.StateBag) {
	// nothing to do
}

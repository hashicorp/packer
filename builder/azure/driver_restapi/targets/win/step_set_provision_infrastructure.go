// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/communicator/azureVmCustomScriptExtension"
	azureservice "github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	storageservice "github.com/mitchellh/packer/builder/azure/driver_restapi/storage_service/request"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"log"
)

type StepSetProvisionInfrastructure struct {
	VmName string
	ServiceName string
	StorageAccountName string
	TempContainerName string
	storageServiceDriver *storageservice.StorageServiceDriver
	flagTempContainerCreated bool
}

func (s *StepSetProvisionInfrastructure) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	reqManager := state.Get(constants.RequestManager).(*azureservice.Manager)

	errorMsg := "Error StepRemoteSession: %s"
	ui.Say("Preparing infrastructure for provision...")

	// get key for storage account
	ui.Message("Getting key for storage account...")
	storageAccountName := s.StorageAccountName
	requestData := reqManager.GetStorageAccountKeys(storageAccountName)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	storageService, err := response.ParseStorageService(resp.Body)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("storageService: %v\n\n", storageService)

	storageAccountKey := storageService.StorageServiceKeys.Primary

	//create storage driver
	storageServiceDriver := storageservice.NewStorageServiceDriver(storageAccountName, storageAccountKey)
	s.storageServiceDriver = storageServiceDriver

	//create temporary container
	s.flagTempContainerCreated = false

	ui.Message("Creating Azure temporary container...")
	_, err = storageServiceDriver.CreateContainer(s.TempContainerName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.flagTempContainerCreated = true

	isOSImage := state.Get(constants.IsOSImage).(bool)

	comm, err := azureVmCustomScriptExtension.New(
		&azureVmCustomScriptExtension.Config{
			ServiceName: s.ServiceName,
			VmName: s.VmName,
			StorageServiceDriver : storageServiceDriver,
			AzureServiceRequestManager : reqManager,
			ContainerName : s.TempContainerName,
			Ui: ui,
			IsOSImage : isOSImage,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	packerCommunicator := packer.Communicator(comm)

	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetProvisionInfrastructure) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Cleaning up infrastructure for provision...")

	if s.flagTempContainerCreated {
		ui.Message("Removing Azure temporary container...")

		_, err := s.storageServiceDriver.DeleteContainer(s.TempContainerName)
		if err != nil {
			ui.Message("Error removing temporary container: " + err.Error())
		}
	}
}

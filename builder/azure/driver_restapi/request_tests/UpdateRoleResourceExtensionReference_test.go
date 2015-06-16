// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"fmt"
	"encoding/base64"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
)

func _TestUpdateRoleResourceExtensionReference(t *testing.T) {

	errMassage := "TestUpdateRoleResourceExtensionReference: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	storageAccountName := "packervhds"
	requestData := reqManager.GetStorageAccountKeys(storageAccountName)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	storageService, err := response.ParseStorageService(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("storageService: %v\n\n", storageService)

	storageAccountKey := storageService.StorageServiceKeys.Primary

	requestData = reqManager.ListResourceExtensions()
	resp, err = reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	list, err := response.ParseResourceExtensionList(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	ext := list.FirstOrNull("CustomScriptExtension")
	fmt.Printf("CustomScriptExtension: %v\n\n", ext)

	if ext == nil {
		t.Errorf(errMassage, "CustomScriptExtension is nil")
	}

	serviceName := "shchTmp"
	vmName := "shchTmp"
	nameOfReference := "PackerCSE"
	nameOfPublisher := ext.Publisher
	nameOfExtension := ext.Name
	versionOfExtension := ext.Version
	state := "enable"

	account := "{\"storageAccountName\":\"" + storageAccountName + "\",\"storageAccountKey\": \"" + storageAccountKey + "\"}";
	runScript := "hello.ps1"
//https://packervhds.blob.core.windows.net/vhds/hello.ps1
	uri := fmt.Sprintf("https://%s.blob.core.windows.net/vhds/%s", storageAccountName, runScript)
	scriptfile := "{\"fileUris\": [\"" + uri + "\"], \"commandToExecute\":\"powershell -ExecutionPolicy Unrestricted -file " + runScript + "\"}"

	params := []request.ResourceExtensionParameterValue {
		request.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPublicConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(scriptfile)),
			Type: "Public",
		},
		request.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPrivateConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(account)),
			Type: "Private",
		},
	}

	requestData = reqManager.UpdateRoleResourceExtensionReference(serviceName, vmName, nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	err = reqManager.ExecuteSync(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData = reqManager.GetDeployment(serviceName, vmName)
	resp, err = reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	deployment, err := response.ParseDeployment(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("\ndeployment:\n\n %v", deployment.RoleInstanceList[0].ResourceExtensionStatusList[1])

	t.Error("eom")
}

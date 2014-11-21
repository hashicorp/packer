// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	"fmt"
)

const extPublisher = "Microsoft.Compute"
const extName = "CustomScriptExtension"
const extRefName = "CustomScriptExtension"


func _TestGetDeploymet(t *testing.T) {

	errMassage := "TestGetDeploymet: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	serviceName := "pkrsrvkqtw0mqcm4"
	vmName := "PkrVMkqtw0mqcm4"



	requestData := reqManager.GetDeployment(serviceName, vmName)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}
	deployment, err := response.ParseDeployment(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	nameOfReference := extRefName
	nameOfPublisher := extPublisher
	nameOfExtension := extName
	versionOfExtension := "1.1"

	var params []request.ResourceExtensionParameterValue

	state := "uninstall"

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


	deployment, err = response.ParseDeployment(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("\ndeployment:\n\n %v", deployment.RoleInstanceList[0].GuestAgentStatus)

	t.Error("eom")
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"fmt"
)

func _TestCreateVirtualMachineDeployment(t *testing.T) {

	errMassage := "TestCreateVirtualMachineDeployment: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	serviceName := "PkrSrvmosl521tfw"
	vmName := "packerVm100"
	vmSize := "Small"
//	certThumbprint := "86E8759D962F983AC32D3F3CCAC941E1B81CAE31"
	userName := "packer"
	userPassword := "Vbnm1234"
	osImageName := "a699494373c04fc0bc8f2bb1389d6106__Windows-Server-2012-R2-201407.01-en.us-127GB.vhd"
	mediaLoc := fmt.Sprintf("https://packervhds.blob.core.windows.net/vhds/%s.vhd", vmName)

	isOSImage := true;

	requestData := reqManager.CreateVirtualMachineDeploymentWin(isOSImage, serviceName, vmName, vmSize, /*certThumbprint,*/ userName, userPassword, osImageName, mediaLoc)
	_, err = reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	t.Error("eom")
}

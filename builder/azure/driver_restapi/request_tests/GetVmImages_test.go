// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response/model"
	"fmt"
)

func _TestGetVmImages(t *testing.T) {

	errMassage := "TestGetVmImages: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	list, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("vmImageList:\n\n")
	model.PrintVmImages(list.VMImages)

	userImageName := "CoreOS"

	fmt.Printf("Image named %s:\n", userImageName)
	first := list.First(userImageName)
	if first != nil {
		t.Logf("%v\n\n", first)
	} else {
		t.Logf("%v\n\n", "Not found!")
	}

	t.Error("eom")
}

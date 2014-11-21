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

func _TestGetOsImages(t *testing.T) {

	errMassage := "GetOsImages: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData := reqManager.GetOsImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	list, err := response.ParseOsImageList(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("osImageList:\n\n")
	model.PrintOsImages(list.OSImages)

	label := "CoreOS"
	location := "West US"

	filteredImageList := list.Filter(label, location)
	list.SortByDateDesc(filteredImageList)
	fmt.Printf("Filtered and Sorted ----------------------------------:\n\n")

	model.PrintOsImages(filteredImageList)



	t.Error("eom")
}

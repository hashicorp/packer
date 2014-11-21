// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"fmt"
)

func _TestListResourceExtensions(t *testing.T) {

	errMassage := "TestListResourceExtensions: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData := reqManager.ListResourceExtensions()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	list, err := response.ParseResourceExtensionList(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("ResourceExtensionList:\n\n")

	for _, val := range(list.ResourceExtensions){
//		fmt.Printf("Name: %s\nLabel: %s\nDescription: %s\n\n\n", val.Name, val.Label, val.Description)
		fmt.Printf("CustomScriptExtension:\n%v\n\n\n", val)
	}

	ext := list.FirstOrNull("CustomScriptExtension")
	fmt.Printf("CustomScriptExtension: %v\n\n", ext)


	t.Error("eom")
}

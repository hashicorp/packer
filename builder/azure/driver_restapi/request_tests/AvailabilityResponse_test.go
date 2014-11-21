// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
)

func _TestAvailabilityResponse(t *testing.T) {
	t.Log("+++TestAvailabilityResponse")
	errMassage := "AvailabilityResponse: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData := reqManager.CheckStorageAccountNameAvailability("ubuntupuppetvhds1")
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	availabilityResponse, err := response.ParseAvailabilityResponse(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	t.Logf("availabilityResponse:\n %v", availabilityResponse)
	t.Error("eom")
}

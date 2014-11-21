// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"testing"
	"time"
	"fmt"
)

func _TestCaptureVmImage(t *testing.T) {

	errMassage := "TestCaptureVmImage: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	now := time.Now()
	label := "PackerMadeUbuntuServer14"

	y,m,d := now.Date()

	tmpServiceName 		:= g_tmpServiceName
	tmpVmName 			:= g_tmpVmName
	userImageName 		:= fmt.Sprintf("%s_%v-%v-%v_%v-%v",label,  y,m,d, now.Hour(), now.Minute() )
	userImageLabel 		:= "packer made image"
	description 		:= "packer made image"
	imageFamily 		:= "PackerMade"
	recommendedVMSize 	:= "Small"

	requestData := reqManager.CaptureVMImage(tmpServiceName, tmpVmName, userImageName, userImageLabel, description, imageFamily, recommendedVMSize )
	_, err = reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}
}

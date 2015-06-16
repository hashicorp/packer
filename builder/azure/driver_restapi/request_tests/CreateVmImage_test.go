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

func _TestCreateVmImage(t *testing.T) {

	errMassage := "TestCreateVmImage: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	now := time.Now()
	label := "TestCreateVmImage"

	mediaLoc := "https://packervhds.blob.core.windows.net/from-user-im/guflyify.qew201409241811330266.vhd"
	os := "Windows"
	y,m,d := now.Date()

	userImageName 		:= fmt.Sprintf("%s_%v-%v-%v_%v-%v",label,  y,m,d, now.Hour(), now.Minute() )
	userImageLabel 		:= "paker made image label"
	description 		:= "paker made image description"
	imageFamily 		:= "TestCreateVmImage"
	recommendedVMSize 	:= "Small"

	const dateLayout = "Mon, 02 Jan 2006 15:04:05 GMT"
	pt := time.Now().UTC()
	publishedDate 	:= pt.Format(dateLayout)

	requestData := reqManager.CreateVMImage(mediaLoc, os,  userImageName, userImageLabel, description, imageFamily, recommendedVMSize, publishedDate )
	err = reqManager.ExecuteSync(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}
}

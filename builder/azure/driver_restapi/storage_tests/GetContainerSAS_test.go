// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package storage_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/storage_service/request"
	"fmt"
	"time"
)

func _TestGetContainerSas(t *testing.T) {

	errMassage := "TestGetContainerSas: %s\n"

	sa := request.NewStorageServiceDriver(g_accountName, g_secret)


	ts := time.Now().UTC()
	fmt.Println("ts: " + ts.String())
	te := ts.Add(time.Hour*24)
	fmt.Println("te: " + te.String())

	signedstart := ts.Format(time.RFC3339)
	fmt.Println("signedstart: " + signedstart)

	containerName := "images"
	sas, err := sa.GetContainerSAS(containerName)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Println("sas: " + sas)

	t.Error("eom")
}

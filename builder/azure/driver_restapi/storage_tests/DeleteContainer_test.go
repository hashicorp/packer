// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package storage_tests

import (
	"testing"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/storage_service/request"
)

func _TestDeleteContainer(t *testing.T) {

	errMassage := "TestDeleteContainer: %s\n"

	sa := request.NewStorageServiceDriver(g_accountName, g_secret)

	containerName := "scch1"
	_, err := sa.DeleteContainer(containerName)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	t.Error("eom")
}

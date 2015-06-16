// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"fmt"
	"testing"
)

func _TestCleanUp(t *testing.T) {

	fmt.Println("TestCleanup...")

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf("Error creating request manager: %s\n", err.Error())
	}

	tmpServiceName := "shchremoveme"
	requestData := reqManager.DeleteCloudServiceAndMedia(tmpServiceName)

	err = reqManager.ExecuteSync(requestData)

	if err != nil {
		t.Errorf("exit with error: %s", err.Error())
		return
	}

	t.Error("eom")

}

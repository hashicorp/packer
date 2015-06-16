// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/driver"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
	"runtime"
	"github.com/mitchellh/packer/builder/azure/driver_restapi"
	"fmt"
)

const (
	psPathLin = "/home/azure/publish_setttings/ps.publishsettings"
	psPathWin = "d:\\Packer.io\\PackerLinux\\ps.publishsettings"
	subscriptionName = "EduardK-Docker"
)

var g_reqManager *request.Manager

func getRequestManager() (*request.Manager, error) {

	if g_reqManager != nil {
		return g_reqManager, nil
	}

	var d driver.IDriverRest
	var err error

	var psPath string

	if runtime.GOOS == constants.Linux {
		psPath = psPathLin
	} else if runtime.GOOS == constants.Windows {
		psPath = psPathWin
	}

	subscriptionInfo, err := driver_restapi.ParsePublishSettings(psPath, subscriptionName)

	if err != nil {
		return nil, fmt.Errorf("ParsePublishSettings error: %s\n", err.Error())
	}

	fmt.Println("id: " + subscriptionInfo.Id)

	d, err = driver.NewTlsDriver(subscriptionInfo.CertData)

	if err != nil {
		return nil, fmt.Errorf("NewTlsDriver error: %s\n", err.Error())
	}

	g_reqManager = &request.Manager{
		SubscrId: subscriptionInfo.Id,
		Driver : d,
		}

	return g_reqManager, err
}

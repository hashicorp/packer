// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
)

func (m *Manager) DeleteDeployment(serviceName, vmName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s", m.SubscrId, serviceName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2009-10-01",
	}

	data := &Data{
		Verb: "DELETE",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
// the operating system disk, attached data disks, and the source blobs for the disks should also be deleted from storage.
func (m *Manager) DeleteDeploymentAndMedia(serviceName, vmName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s?comp=media", m.SubscrId, serviceName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2013-08-01",
	}

	data := &Data{
		Verb: "DELETE",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}

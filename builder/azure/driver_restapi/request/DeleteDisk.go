// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
)

func (m *Manager) DeleteDisk(diskName string) (*Data) {
	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/disks/%s", m.SubscrId, diskName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2012-08-01",
	}

	data := &Data{
		Verb: "DELETE",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
//the blob that is associated with the disk is also deleted
func (m *Manager) DeleteDiskAndMedia(diskName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/disks/%s?comp=media", m.SubscrId, diskName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2012-08-01",
	}

	data := &Data{
		Verb: "DELETE",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}


// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
)

func (m *Manager) DeleteCloudService(serviceName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s", m.SubscrId, serviceName)

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

func (m *Manager) DeleteCloudServiceAndMedia(serviceName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s?comp=media", m.SubscrId, serviceName)

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

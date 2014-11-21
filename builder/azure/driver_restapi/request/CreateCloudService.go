// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"encoding/base64"
	"fmt"
	"bytes"
)

func (m *Manager) CreateCloudService(serviceName, location string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices", m.SubscrId)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2010-10-28",
	}

	serviceNameLabel := base64.StdEncoding.EncodeToString([]byte(serviceName))

	var buff bytes.Buffer
	buff.WriteString("<?xml version='1.0' encoding='utf-8'?>")
	buff.WriteString("<CreateHostedService xmlns='http://schemas.microsoft.com/windowsazure'>")
	buff.WriteString("<ServiceName>" +serviceName + "</ServiceName>")
	buff.WriteString("<Label>" + serviceNameLabel +"</Label>")
	buff.WriteString("<Description></Description>")
	buff.WriteString("<Location>" + location + "</Location>")
	buff.WriteString("</CreateHostedService>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}


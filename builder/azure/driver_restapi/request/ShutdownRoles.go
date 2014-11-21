// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) ShutdownRoles(serviceName string, vmName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s/roles/Operations", m.SubscrId, serviceName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2013-06-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<ShutdownRolesOperation xmlns='http://schemas.microsoft.com/windowsazure' xmlns:i='http://www.w3.org/2001/XMLSchema-instance'>")
	buff.WriteString("<OperationType>ShutdownRolesOperation</OperationType>")
	buff.WriteString("<Roles>")
	buff.WriteString("<Name>" +vmName + "</Name>")
	buff.WriteString("</Roles>")
	buff.WriteString("<PostShutdownAction>StoppedDeallocated</PostShutdownAction>")
	buff.WriteString("</ShutdownRolesOperation>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}

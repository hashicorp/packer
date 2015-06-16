// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

type ResourceExtensionParameterValue struct {
	Key string
	Value string
	Type string
}

func (m *Manager) UpdateRoleResourceExtensionReference(serviceName, vmName, nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state string, params []ResourceExtensionParameterValue) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s/roles/%s", m.SubscrId, serviceName, vmName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-06-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<PersistentVMRole xmlns='http://schemas.microsoft.com/windowsazure' xmlns:i='http://www.w3.org/2001/XMLSchema-instance'>")

	buff.WriteString("<RoleName>"+ vmName +"</RoleName>")
	buff.WriteString("<RoleType>PersistentVMRole</RoleType>")

	buff.WriteString("<ResourceExtensionReferences>")
		buff.WriteString("<ResourceExtensionReference>")
			buff.WriteString("<ReferenceName>"+ nameOfReference +"</ReferenceName>")
			buff.WriteString("<Publisher>"+ nameOfPublisher +"</Publisher>")
			buff.WriteString("<Name>"+ nameOfExtension +"</Name>")
			buff.WriteString("<Version>"+ versionOfExtension +"</Version>")
	if len(params) > 0 {
				buff.WriteString("<ResourceExtensionParameterValues>")
		for _, param := range(params) {
						buff.WriteString("<ResourceExtensionParameterValue>")
							buff.WriteString("<Key>"+ param.Key +"</Key>")
							buff.WriteString("<Value>"+ param.Value +"</Value>")
							buff.WriteString("<Type>"+ param.Type +"</Type>")
						buff.WriteString("</ResourceExtensionParameterValue>")
		}
				buff.WriteString("</ResourceExtensionParameterValues>")
	}
			buff.WriteString("<State>"+ state +"</State>")
		buff.WriteString("</ResourceExtensionReference>")
	buff.WriteString("</ResourceExtensionReferences>")

	buff.WriteString("<ProvisionGuestAgent>true</ProvisionGuestAgent>")
	buff.WriteString("</PersistentVMRole>")

	data := &Data{
		Verb: "PUT",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}

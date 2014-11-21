// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) CaptureVMImage(serviceName, vmName, name, label, description, imageFamily, recommendedVMSize string) (*Data) {
	language := "english"
	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s/roleinstances/%s/Operations", m.SubscrId, serviceName, vmName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-05-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<CaptureRoleAsVMImageOperation xmlns='http://schemas.microsoft.com/windowsazure' xmlns:i='http://www.w3.org/2001/XMLSchema-instance'>")
	buff.WriteString("<OperationType>CaptureRoleAsVMImageOperation</OperationType>")
	buff.WriteString("<OSState>Generalized</OSState>")
	buff.WriteString("<VMImageName>"+ name +"</VMImageName>")
	buff.WriteString("<VMImageLabel>"+ label +"</VMImageLabel>")
	buff.WriteString("<Description>"+ description +"</Description>")
	buff.WriteString("<Language>"+ language +"</Language>")
	buff.WriteString("<ImageFamily>"+ imageFamily +"</ImageFamily>")
	buff.WriteString("<RecommendedVMSize>"+ recommendedVMSize +"</RecommendedVMSize>")
	buff.WriteString("</CaptureRoleAsVMImageOperation>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}


// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) AddCertificate(serviceName, certDataBase64, certFormat, password string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/certificates", m.SubscrId, serviceName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2009-10-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<?xml version='1.0' encoding='utf-8'?>")
	buff.WriteString("<CertificateFile xmlns='http://schemas.microsoft.com/windowsazure'>")
	buff.WriteString("<Data>" + certDataBase64 + "</Data>")
	buff.WriteString("<CertificateFormat>" + certFormat +"</CertificateFormat>")
	buff.WriteString("<Password>" + password + "</Password>")
	buff.WriteString("</CertificateFile>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
)

type ServiceCertificateList struct {
	XMLName   xml.Name `xml:"Certificates"`
	Certificates []Certificate `xml:"Certificate"`
}

type Certificate struct {
	CertificateUrl string
	Thumbprint string
	ThumbprintAlgorithm string
	Data string
}

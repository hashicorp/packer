// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) CreateVMImage(mediaLoc, os, name, label, description, imageFamily, recommendedVMSize, publishedDate string) (*Data) {
	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/vmimages", m.SubscrId)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-05-01",
	}

//	language := "english"

	var buff bytes.Buffer
	buff.WriteString("<VMImages xmlns='http://schemas.microsoft.com/windowsazure' xmlns:i='http://www.w3.org/2001/XMLSchema-instance'>")
		buff.WriteString("<VMImage>")
			buff.WriteString("<Name>"+ "shchName" +"</Name>")
			buff.WriteString("<Label>"+ "shchLabel" +"</Label>")
//			buff.WriteString("<Description>"+ description +"</Description>")
			buff.WriteString("<OSDiskConfiguration>")
//				buff.WriteString("<HostCaching>ReadWrite</HostCaching>")
				buff.WriteString("<OSState>Generalized</OSState>")
				buff.WriteString("<OS>Windows</OS>")
				buff.WriteString("<MediaLink>https://packervhds.blob.core.windows.net/from-user-im/guflyify.qew201409241811330266.vhd</MediaLink>")
			buff.WriteString("</OSDiskConfiguration>")
//			buff.WriteString("<DataDiskConfigurations><DataDiskConfiguration/></DataDiskConfigurations>")
//			buff.WriteString("<Language>"+ language +"</Language>")
//			buff.WriteString("<ImageFamily>"+ imageFamily +"</ImageFamily>")
//			buff.WriteString("<RecommendedVMSize>"+ recommendedVMSize +"</RecommendedVMSize>")
//			buff.WriteString("<Eula></Eula>")
//			buff.WriteString("<IconUri></IconUri>")
//			buff.WriteString("<SmallIconUri></SmallIconUri>")
//			buff.WriteString("<PrivacyUri></PrivacyUri>")
//			buff.WriteString("<PublishedDate>"+ publishedDate +"</PublishedDate>")
//			buff.WriteString("<ShowInGui>true</ShowInGui>")
		buff.WriteString("</VMImage>")
	buff.WriteString("</VMImages>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}


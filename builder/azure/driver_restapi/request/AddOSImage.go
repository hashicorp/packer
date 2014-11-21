// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) AddOSImage(mediaLoc, os, name, label, description, imageFamily, recommendedVMSize, publishedDate string) (*Data) {
	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/images", m.SubscrId)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-05-01",
	}

	language := "english"

	var buff bytes.Buffer
	buff.WriteString("<OSImage xmlns='http://schemas.microsoft.com/windowsazure' xmlns:i='http://www.w3.org/2001/XMLSchema-instance'>")
		buff.WriteString("<Label>shchLabel</Label>")
		buff.WriteString("<Label>"+ label +"</Label>")
		buff.WriteString("<MediaLink>"+ mediaLoc +"</MediaLink>")
		buff.WriteString("<Name>"+ name +"</Name>")
		buff.WriteString("<OS>"+ os +"</OS>")
//		buff.WriteString("<Eula></Eula>")
		buff.WriteString("<Description>"+ description +"</Description>")
		buff.WriteString("<ImageFamily>"+ imageFamily +"</ImageFamily>")
		buff.WriteString("<PublishedDate>"+ publishedDate +"</PublishedDate>")
		buff.WriteString("<ShowInGui>true</ShowInGui>")
//		buff.WriteString("<PrivacyUri></PrivacyUri>")
//		buff.WriteString("<IconUri></IconUri>")
		buff.WriteString("<RecommendedVMSize>"+ recommendedVMSize +"</RecommendedVMSize>")
//		buff.WriteString("<SmallIconUri></SmallIconUri>")
		buff.WriteString("<Language>"+ language +"</Language>")
	buff.WriteString("</OSImage>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}


// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
	"regexp"
)

type ResourceExtensionList struct {
	XMLName   xml.Name `xml:"ResourceExtensions"`
	ResourceExtensions []ResourceExtension `xml:"ResourceExtension"`
}

type ResourceExtension struct {
	Publisher string
	Name string
	Version string
	Label string
	Description string
	PublicConfigurationSchema string
	PrivateConfigurationSchema string
	SampleConfig string
	ReplicationCompleted string
	Eula string
	PrivacyUri string
	HomepageUri string
	IsJsonExtension string
	IsInternalExtension string
	DisallowMajorVersionUpgrade string
	CompanyName string
	SupportedOS string
	PublishedDate string
}

func (l *ResourceExtensionList) FirstOrNull(name string) *ResourceExtension {
	pattern := name
	for _, re := range(l.ResourceExtensions){
		matchName, _ := regexp.MatchString(pattern, re.Name)
		if( matchName ) {
			return &re
		}
	}

	return nil
}

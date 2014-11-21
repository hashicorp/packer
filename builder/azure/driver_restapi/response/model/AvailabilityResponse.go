// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import "encoding/xml"

type AvailabilityResponse struct {
	XMLName   			xml.Name 	`xml:"AvailabilityResponse"`
	Xmlns	  			string 		`xml:"xmlns,attr"`
	Result 				string
	Reason 				string
}

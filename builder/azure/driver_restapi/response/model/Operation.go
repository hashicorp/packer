// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import "encoding/xml"

type Operation struct {
	XMLName   			xml.Name 	`xml:"Operation"`
	Xmlns	  			string 		`xml:"xmlns,attr"`
	ID 					string
	Status 				string
	HttpStatusCode 		string
	Error 				Error 		`xml:"Error"`
}

type Error struct {
	Code 		string
	Message 	string
}

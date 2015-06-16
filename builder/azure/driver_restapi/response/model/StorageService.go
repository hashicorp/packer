// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model


type StorageService struct {
	Url 					string
	StorageServiceKeys 		StorageServiceKeys
}

type StorageServiceKeys struct {
	Primary 		string
	Secondary 		string
}



// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateDbHomeDetails The representation of CreateDbHomeDetails
type CreateDbHomeDetails struct {
	Database *CreateDatabaseDetails `mandatory:"true" json:"database"`

	// A valid Oracle database version. To get a list of supported versions, use the ListDbVersions operation.
	DbVersion *string `mandatory:"true" json:"dbVersion"`

	// The user-provided name of the database home.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateDbHomeDetails) String() string {
	return common.PointerString(m)
}

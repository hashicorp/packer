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

// CreateBackupDetails The representation of CreateBackupDetails
type CreateBackupDetails struct {

	// The OCID of the database.
	DatabaseId *string `mandatory:"true" json:"databaseId"`

	// The user-friendly name for the backup. It does not have to be unique.
	DisplayName *string `mandatory:"true" json:"displayName"`
}

func (m CreateBackupDetails) String() string {
	return common.PointerString(m)
}

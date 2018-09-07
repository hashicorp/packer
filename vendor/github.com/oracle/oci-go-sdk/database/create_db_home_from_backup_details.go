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

// CreateDbHomeFromBackupDetails The representation of CreateDbHomeFromBackupDetails
type CreateDbHomeFromBackupDetails struct {
	Database *CreateDatabaseFromBackupDetails `mandatory:"true" json:"database"`

	// The user-provided name of the database home.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateDbHomeFromBackupDetails) String() string {
	return common.PointerString(m)
}

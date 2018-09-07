// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// File Storage Service API
//
// The API for the File Storage Service.
//

package filestorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateExportDetails The representation of CreateExportDetails
type CreateExportDetails struct {

	// The OCID of this export's export set.
	ExportSetId *string `mandatory:"true" json:"exportSetId"`

	// The OCID of this export's file system.
	FileSystemId *string `mandatory:"true" json:"fileSystemId"`

	// Path used to access the associated file system.
	// Avoid entering confidential information.
	// Example: `/mediafiles`
	Path *string `mandatory:"true" json:"path"`
}

func (m CreateExportDetails) String() string {
	return common.PointerString(m)
}

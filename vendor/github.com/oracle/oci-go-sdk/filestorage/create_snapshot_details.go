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

// CreateSnapshotDetails The representation of CreateSnapshotDetails
type CreateSnapshotDetails struct {

	// The OCID of this export's file system.
	FileSystemId *string `mandatory:"true" json:"fileSystemId"`

	// Name of the snapshot. This value is immutable. It must also be unique with respect
	// to all other non-DELETED snapshots on the associated file
	// system.
	// Avoid entering confidential information.
	// Example: `Sunday`
	Name *string `mandatory:"true" json:"name"`
}

func (m CreateSnapshotDetails) String() string {
	return common.PointerString(m)
}

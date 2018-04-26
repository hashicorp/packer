// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateVolumeBackupDetails The representation of UpdateVolumeBackupDetails
type UpdateVolumeBackupDetails struct {

	// A friendly user-specified name for the volume backup.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m UpdateVolumeBackupDetails) String() string {
	return common.PointerString(m)
}

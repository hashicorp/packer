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

// CreateVolumeBackupDetails The representation of CreateVolumeBackupDetails
type CreateVolumeBackupDetails struct {

	// The OCID of the volume that needs to be backed up.
	VolumeId *string `mandatory:"true" json:"volumeId"`

	// A user-friendly name for the volume backup. Does not have to be unique and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateVolumeBackupDetails) String() string {
	return common.PointerString(m)
}

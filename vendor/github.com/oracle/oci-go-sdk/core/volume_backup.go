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

// VolumeBackup A point-in-time copy of a volume that can then be used to create a new block volume
// or recover a block volume. For more information, see
// Overview of Cloud Volume Storage (https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type VolumeBackup struct {

	// The OCID of the compartment that contains the volume backup.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name for the volume backup. Does not have to be unique and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the volume backup.
	Id *string `mandatory:"true" json:"id"`

	// The current state of a volume backup.
	LifecycleState VolumeBackupLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the volume backup was created. This is the time the actual point-in-time image
	// of the volume data was taken. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The size of the volume, in GBs.
	SizeInGBs *int `mandatory:"false" json:"sizeInGBs"`

	// The size of the volume in MBs. The value must be a multiple of 1024.
	// This field is deprecated. Please use sizeInGBs.
	SizeInMBs *int `mandatory:"false" json:"sizeInMBs"`

	// The date and time the request to create the volume backup was received. Format defined by RFC3339.
	TimeRequestReceived *common.SDKTime `mandatory:"false" json:"timeRequestReceived"`

	// The size used by the backup, in GBs. It is typically smaller than sizeInGBs, depending on the space
	// consumed on the volume and whether the backup is full or incremental.
	UniqueSizeInGBs *int `mandatory:"false" json:"uniqueSizeInGBs"`

	// The size used by the backup, in MBs. It is typically smaller than sizeInMBs, depending on the space
	// consumed on the volume and whether the backup is full or incremental.
	// This field is deprecated. Please use uniqueSizeInGBs.
	UniqueSizeInMbs *int `mandatory:"false" json:"uniqueSizeInMbs"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"false" json:"volumeId"`
}

func (m VolumeBackup) String() string {
	return common.PointerString(m)
}

// VolumeBackupLifecycleStateEnum Enum with underlying type: string
type VolumeBackupLifecycleStateEnum string

// Set of constants representing the allowable values for VolumeBackupLifecycleState
const (
	VolumeBackupLifecycleStateCreating        VolumeBackupLifecycleStateEnum = "CREATING"
	VolumeBackupLifecycleStateAvailable       VolumeBackupLifecycleStateEnum = "AVAILABLE"
	VolumeBackupLifecycleStateTerminating     VolumeBackupLifecycleStateEnum = "TERMINATING"
	VolumeBackupLifecycleStateTerminated      VolumeBackupLifecycleStateEnum = "TERMINATED"
	VolumeBackupLifecycleStateFaulty          VolumeBackupLifecycleStateEnum = "FAULTY"
	VolumeBackupLifecycleStateRequestReceived VolumeBackupLifecycleStateEnum = "REQUEST_RECEIVED"
)

var mappingVolumeBackupLifecycleState = map[string]VolumeBackupLifecycleStateEnum{
	"CREATING":         VolumeBackupLifecycleStateCreating,
	"AVAILABLE":        VolumeBackupLifecycleStateAvailable,
	"TERMINATING":      VolumeBackupLifecycleStateTerminating,
	"TERMINATED":       VolumeBackupLifecycleStateTerminated,
	"FAULTY":           VolumeBackupLifecycleStateFaulty,
	"REQUEST_RECEIVED": VolumeBackupLifecycleStateRequestReceived,
}

// GetVolumeBackupLifecycleStateEnumValues Enumerates the set of values for VolumeBackupLifecycleState
func GetVolumeBackupLifecycleStateEnumValues() []VolumeBackupLifecycleStateEnum {
	values := make([]VolumeBackupLifecycleStateEnum, 0)
	for _, v := range mappingVolumeBackupLifecycleState {
		values = append(values, v)
	}
	return values
}

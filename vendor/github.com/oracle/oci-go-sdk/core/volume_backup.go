// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// VolumeBackup A point-in-time copy of a volume that can then be used to create a new block volume
// or recover a block volume. For more information, see
// Overview of Cloud Volume Storage (https://docs.cloud.oracle.com/Content/Block/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
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
	// of the volume data was taken. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The type of a volume backup.
	Type VolumeBackupTypeEnum `mandatory:"true" json:"type"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// System tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// The date and time the volume backup will expire and be automatically deleted.
	// Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339). This parameter will always be present for backups that
	// were created automatically by a scheduled-backup policy. For manually created backups,
	// it will be absent, signifying that there is no expiration time and the backup will
	// last forever until manually deleted.
	ExpirationTime *common.SDKTime `mandatory:"false" json:"expirationTime"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID of the Key Management key which is the master encryption key for the volume backup.
	// For more information about the Key Management service and encryption keys, see
	// Overview of Key Management (https://docs.cloud.oracle.com/Content/KeyManagement/Concepts/keyoverview.htm) and
	// Using Keys (https://docs.cloud.oracle.com/Content/KeyManagement/Tasks/usingkeys.htm).
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// The size of the volume, in GBs.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	// The size of the volume in MBs. The value must be a multiple of 1024.
	// This field is deprecated. Please use sizeInGBs.
	SizeInMBs *int64 `mandatory:"false" json:"sizeInMBs"`

	// Specifies whether the backup was created manually, or via scheduled backup policy.
	SourceType VolumeBackupSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	// The OCID of the source volume backup.
	SourceVolumeBackupId *string `mandatory:"false" json:"sourceVolumeBackupId"`

	// The date and time the request to create the volume backup was received. Format defined by [RFC3339]https://tools.ietf.org/html/rfc3339.
	TimeRequestReceived *common.SDKTime `mandatory:"false" json:"timeRequestReceived"`

	// The size used by the backup, in GBs. It is typically smaller than sizeInGBs, depending on the space
	// consumed on the volume and whether the backup is full or incremental.
	UniqueSizeInGBs *int64 `mandatory:"false" json:"uniqueSizeInGBs"`

	// The size used by the backup, in MBs. It is typically smaller than sizeInMBs, depending on the space
	// consumed on the volume and whether the backup is full or incremental.
	// This field is deprecated. Please use uniqueSizeInGBs.
	UniqueSizeInMbs *int64 `mandatory:"false" json:"uniqueSizeInMbs"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"false" json:"volumeId"`
}

func (m VolumeBackup) String() string {
	return common.PointerString(m)
}

// VolumeBackupLifecycleStateEnum Enum with underlying type: string
type VolumeBackupLifecycleStateEnum string

// Set of constants representing the allowable values for VolumeBackupLifecycleStateEnum
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

// GetVolumeBackupLifecycleStateEnumValues Enumerates the set of values for VolumeBackupLifecycleStateEnum
func GetVolumeBackupLifecycleStateEnumValues() []VolumeBackupLifecycleStateEnum {
	values := make([]VolumeBackupLifecycleStateEnum, 0)
	for _, v := range mappingVolumeBackupLifecycleState {
		values = append(values, v)
	}
	return values
}

// VolumeBackupSourceTypeEnum Enum with underlying type: string
type VolumeBackupSourceTypeEnum string

// Set of constants representing the allowable values for VolumeBackupSourceTypeEnum
const (
	VolumeBackupSourceTypeManual    VolumeBackupSourceTypeEnum = "MANUAL"
	VolumeBackupSourceTypeScheduled VolumeBackupSourceTypeEnum = "SCHEDULED"
)

var mappingVolumeBackupSourceType = map[string]VolumeBackupSourceTypeEnum{
	"MANUAL":    VolumeBackupSourceTypeManual,
	"SCHEDULED": VolumeBackupSourceTypeScheduled,
}

// GetVolumeBackupSourceTypeEnumValues Enumerates the set of values for VolumeBackupSourceTypeEnum
func GetVolumeBackupSourceTypeEnumValues() []VolumeBackupSourceTypeEnum {
	values := make([]VolumeBackupSourceTypeEnum, 0)
	for _, v := range mappingVolumeBackupSourceType {
		values = append(values, v)
	}
	return values
}

// VolumeBackupTypeEnum Enum with underlying type: string
type VolumeBackupTypeEnum string

// Set of constants representing the allowable values for VolumeBackupTypeEnum
const (
	VolumeBackupTypeFull        VolumeBackupTypeEnum = "FULL"
	VolumeBackupTypeIncremental VolumeBackupTypeEnum = "INCREMENTAL"
)

var mappingVolumeBackupType = map[string]VolumeBackupTypeEnum{
	"FULL":        VolumeBackupTypeFull,
	"INCREMENTAL": VolumeBackupTypeIncremental,
}

// GetVolumeBackupTypeEnumValues Enumerates the set of values for VolumeBackupTypeEnum
func GetVolumeBackupTypeEnumValues() []VolumeBackupTypeEnum {
	values := make([]VolumeBackupTypeEnum, 0)
	for _, v := range mappingVolumeBackupType {
		values = append(values, v)
	}
	return values
}

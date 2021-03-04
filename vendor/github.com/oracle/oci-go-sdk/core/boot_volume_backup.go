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

// BootVolumeBackup A point-in-time copy of a boot volume that can then be used to create
// a new boot volume or recover a boot volume. For more information, see Overview
// of Boot Volume Backups (https://docs.cloud.oracle.com/Content/Block/Concepts/bootvolumebackups.htm)
// To use any of the API operations, you must be authorized in an IAM policy.
// If you're not authorized, talk to an administrator. If you're an administrator
// who needs to write policies to give users access, see Getting Started with
// Policies (https://docs.cloud.oracle.com/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type BootVolumeBackup struct {

	// The OCID of the compartment that contains the boot volume backup.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name for the boot volume backup. Does not have to be unique and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the boot volume backup.
	Id *string `mandatory:"true" json:"id"`

	// The current state of a boot volume backup.
	LifecycleState BootVolumeBackupLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the boot volume backup was created. This is the time the actual point-in-time image
	// of the volume data was taken. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the boot volume.
	BootVolumeId *string `mandatory:"false" json:"bootVolumeId"`

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

	// The image OCID used to create the boot volume the backup is taken from.
	ImageId *string `mandatory:"false" json:"imageId"`

	// The OCID of the Key Management master encryption assigned to the boot volume backup.
	// For more information about the Key Management service and encryption keys, see
	// Overview of Key Management (https://docs.cloud.oracle.com/Content/KeyManagement/Concepts/keyoverview.htm) and
	// Using Keys (https://docs.cloud.oracle.com/Content/KeyManagement/Tasks/usingkeys.htm).
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// The size of the boot volume, in GBs.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	// The OCID of the source boot volume backup.
	SourceBootVolumeBackupId *string `mandatory:"false" json:"sourceBootVolumeBackupId"`

	// Specifies whether the backup was created manually, or via scheduled backup policy.
	SourceType BootVolumeBackupSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	// The date and time the request to create the boot volume backup was received. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeRequestReceived *common.SDKTime `mandatory:"false" json:"timeRequestReceived"`

	// The type of a volume backup.
	Type BootVolumeBackupTypeEnum `mandatory:"false" json:"type,omitempty"`

	// The size used by the backup, in GBs. It is typically smaller than sizeInGBs, depending on the space
	// consumed on the boot volume and whether the backup is full or incremental.
	UniqueSizeInGBs *int64 `mandatory:"false" json:"uniqueSizeInGBs"`
}

func (m BootVolumeBackup) String() string {
	return common.PointerString(m)
}

// BootVolumeBackupLifecycleStateEnum Enum with underlying type: string
type BootVolumeBackupLifecycleStateEnum string

// Set of constants representing the allowable values for BootVolumeBackupLifecycleStateEnum
const (
	BootVolumeBackupLifecycleStateCreating        BootVolumeBackupLifecycleStateEnum = "CREATING"
	BootVolumeBackupLifecycleStateAvailable       BootVolumeBackupLifecycleStateEnum = "AVAILABLE"
	BootVolumeBackupLifecycleStateTerminating     BootVolumeBackupLifecycleStateEnum = "TERMINATING"
	BootVolumeBackupLifecycleStateTerminated      BootVolumeBackupLifecycleStateEnum = "TERMINATED"
	BootVolumeBackupLifecycleStateFaulty          BootVolumeBackupLifecycleStateEnum = "FAULTY"
	BootVolumeBackupLifecycleStateRequestReceived BootVolumeBackupLifecycleStateEnum = "REQUEST_RECEIVED"
)

var mappingBootVolumeBackupLifecycleState = map[string]BootVolumeBackupLifecycleStateEnum{
	"CREATING":         BootVolumeBackupLifecycleStateCreating,
	"AVAILABLE":        BootVolumeBackupLifecycleStateAvailable,
	"TERMINATING":      BootVolumeBackupLifecycleStateTerminating,
	"TERMINATED":       BootVolumeBackupLifecycleStateTerminated,
	"FAULTY":           BootVolumeBackupLifecycleStateFaulty,
	"REQUEST_RECEIVED": BootVolumeBackupLifecycleStateRequestReceived,
}

// GetBootVolumeBackupLifecycleStateEnumValues Enumerates the set of values for BootVolumeBackupLifecycleStateEnum
func GetBootVolumeBackupLifecycleStateEnumValues() []BootVolumeBackupLifecycleStateEnum {
	values := make([]BootVolumeBackupLifecycleStateEnum, 0)
	for _, v := range mappingBootVolumeBackupLifecycleState {
		values = append(values, v)
	}
	return values
}

// BootVolumeBackupSourceTypeEnum Enum with underlying type: string
type BootVolumeBackupSourceTypeEnum string

// Set of constants representing the allowable values for BootVolumeBackupSourceTypeEnum
const (
	BootVolumeBackupSourceTypeManual    BootVolumeBackupSourceTypeEnum = "MANUAL"
	BootVolumeBackupSourceTypeScheduled BootVolumeBackupSourceTypeEnum = "SCHEDULED"
)

var mappingBootVolumeBackupSourceType = map[string]BootVolumeBackupSourceTypeEnum{
	"MANUAL":    BootVolumeBackupSourceTypeManual,
	"SCHEDULED": BootVolumeBackupSourceTypeScheduled,
}

// GetBootVolumeBackupSourceTypeEnumValues Enumerates the set of values for BootVolumeBackupSourceTypeEnum
func GetBootVolumeBackupSourceTypeEnumValues() []BootVolumeBackupSourceTypeEnum {
	values := make([]BootVolumeBackupSourceTypeEnum, 0)
	for _, v := range mappingBootVolumeBackupSourceType {
		values = append(values, v)
	}
	return values
}

// BootVolumeBackupTypeEnum Enum with underlying type: string
type BootVolumeBackupTypeEnum string

// Set of constants representing the allowable values for BootVolumeBackupTypeEnum
const (
	BootVolumeBackupTypeFull        BootVolumeBackupTypeEnum = "FULL"
	BootVolumeBackupTypeIncremental BootVolumeBackupTypeEnum = "INCREMENTAL"
)

var mappingBootVolumeBackupType = map[string]BootVolumeBackupTypeEnum{
	"FULL":        BootVolumeBackupTypeFull,
	"INCREMENTAL": BootVolumeBackupTypeIncremental,
}

// GetBootVolumeBackupTypeEnumValues Enumerates the set of values for BootVolumeBackupTypeEnum
func GetBootVolumeBackupTypeEnumValues() []BootVolumeBackupTypeEnum {
	values := make([]BootVolumeBackupTypeEnum, 0)
	for _, v := range mappingBootVolumeBackupType {
		values = append(values, v)
	}
	return values
}

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

// FileSystemSummary Summary information for a file system.
type FileSystemSummary struct {

	// The number of bytes consumed by the file system, including
	// any snapshots. This number reflects the metered size of the file
	// system and is updated asynchronously with respect to
	// updates to the file system.
	MeteredBytes *int `mandatory:"true" json:"meteredBytes"`

	// The OCID of the compartment that contains the file system.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. It does not have to be unique, and it is changeable.
	// Avoid entering confidential information.
	// Example: `My file system`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the file system.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the file system.
	LifecycleState FileSystemSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the file system was created, expressed
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The availability domain the file system is in. May be unset
	// as a blank or NULL value.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`
}

func (m FileSystemSummary) String() string {
	return common.PointerString(m)
}

// FileSystemSummaryLifecycleStateEnum Enum with underlying type: string
type FileSystemSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for FileSystemSummaryLifecycleState
const (
	FileSystemSummaryLifecycleStateCreating FileSystemSummaryLifecycleStateEnum = "CREATING"
	FileSystemSummaryLifecycleStateActive   FileSystemSummaryLifecycleStateEnum = "ACTIVE"
	FileSystemSummaryLifecycleStateDeleting FileSystemSummaryLifecycleStateEnum = "DELETING"
	FileSystemSummaryLifecycleStateDeleted  FileSystemSummaryLifecycleStateEnum = "DELETED"
)

var mappingFileSystemSummaryLifecycleState = map[string]FileSystemSummaryLifecycleStateEnum{
	"CREATING": FileSystemSummaryLifecycleStateCreating,
	"ACTIVE":   FileSystemSummaryLifecycleStateActive,
	"DELETING": FileSystemSummaryLifecycleStateDeleting,
	"DELETED":  FileSystemSummaryLifecycleStateDeleted,
}

// GetFileSystemSummaryLifecycleStateEnumValues Enumerates the set of values for FileSystemSummaryLifecycleState
func GetFileSystemSummaryLifecycleStateEnumValues() []FileSystemSummaryLifecycleStateEnum {
	values := make([]FileSystemSummaryLifecycleStateEnum, 0)
	for _, v := range mappingFileSystemSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

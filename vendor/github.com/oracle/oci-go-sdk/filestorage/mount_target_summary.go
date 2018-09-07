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

// MountTargetSummary Summary information for the specified mount target.
type MountTargetSummary struct {

	// The OCID of the compartment that contains the mount target.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. It does not have to be unique, and it is changeable.
	// Avoid entering confidential information.
	// Example: `My mount target`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the mount target.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the mount target.
	LifecycleState MountTargetSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCIDs of the private IP addresses associated with this mount target.
	PrivateIpIds []string `mandatory:"true" json:"privateIpIds"`

	// The OCID of the subnet the mount target is in.
	SubnetId *string `mandatory:"true" json:"subnetId"`

	// The date and time the mount target was created, expressed
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The availability domain the mount target is in. May be unset
	// as a blank or NULL value.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID of the associated export set. Controls what file
	// systems will be exported using Network File System (NFS) protocol on
	// this mount target.
	ExportSetId *string `mandatory:"false" json:"exportSetId"`
}

func (m MountTargetSummary) String() string {
	return common.PointerString(m)
}

// MountTargetSummaryLifecycleStateEnum Enum with underlying type: string
type MountTargetSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for MountTargetSummaryLifecycleState
const (
	MountTargetSummaryLifecycleStateCreating MountTargetSummaryLifecycleStateEnum = "CREATING"
	MountTargetSummaryLifecycleStateActive   MountTargetSummaryLifecycleStateEnum = "ACTIVE"
	MountTargetSummaryLifecycleStateDeleting MountTargetSummaryLifecycleStateEnum = "DELETING"
	MountTargetSummaryLifecycleStateDeleted  MountTargetSummaryLifecycleStateEnum = "DELETED"
	MountTargetSummaryLifecycleStateFailed   MountTargetSummaryLifecycleStateEnum = "FAILED"
)

var mappingMountTargetSummaryLifecycleState = map[string]MountTargetSummaryLifecycleStateEnum{
	"CREATING": MountTargetSummaryLifecycleStateCreating,
	"ACTIVE":   MountTargetSummaryLifecycleStateActive,
	"DELETING": MountTargetSummaryLifecycleStateDeleting,
	"DELETED":  MountTargetSummaryLifecycleStateDeleted,
	"FAILED":   MountTargetSummaryLifecycleStateFailed,
}

// GetMountTargetSummaryLifecycleStateEnumValues Enumerates the set of values for MountTargetSummaryLifecycleState
func GetMountTargetSummaryLifecycleStateEnumValues() []MountTargetSummaryLifecycleStateEnum {
	values := make([]MountTargetSummaryLifecycleStateEnum, 0)
	for _, v := range mappingMountTargetSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

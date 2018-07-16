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

// BackupSummary A database backup
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an administrator. If you're an administrator who needs to write policies to give users access, see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type BackupSummary struct {

	// The name of the Availability Domain that the backup is located in.
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The Oracle Database Edition of the DbSystem on which the backup was taken.
	DatabaseEdition *string `mandatory:"false" json:"databaseEdition"`

	// The OCID of the database.
	DatabaseId *string `mandatory:"false" json:"databaseId"`

	// Size of the database in mega-bytes at the time the backup was taken.
	DbDataSizeInMBs *int `mandatory:"false" json:"dbDataSizeInMBs"`

	// The user-friendly name for the backup. It does not have to be unique.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the backup.
	Id *string `mandatory:"false" json:"id"`

	// Additional information about the current lifecycleState.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The current state of the backup.
	LifecycleState BackupSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the backup was completed.
	TimeEnded *common.SDKTime `mandatory:"false" json:"timeEnded"`

	// The date and time the backup starts.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The type of backup.
	Type BackupSummaryTypeEnum `mandatory:"false" json:"type,omitempty"`
}

func (m BackupSummary) String() string {
	return common.PointerString(m)
}

// BackupSummaryLifecycleStateEnum Enum with underlying type: string
type BackupSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for BackupSummaryLifecycleState
const (
	BackupSummaryLifecycleStateCreating  BackupSummaryLifecycleStateEnum = "CREATING"
	BackupSummaryLifecycleStateActive    BackupSummaryLifecycleStateEnum = "ACTIVE"
	BackupSummaryLifecycleStateDeleting  BackupSummaryLifecycleStateEnum = "DELETING"
	BackupSummaryLifecycleStateDeleted   BackupSummaryLifecycleStateEnum = "DELETED"
	BackupSummaryLifecycleStateFailed    BackupSummaryLifecycleStateEnum = "FAILED"
	BackupSummaryLifecycleStateRestoring BackupSummaryLifecycleStateEnum = "RESTORING"
)

var mappingBackupSummaryLifecycleState = map[string]BackupSummaryLifecycleStateEnum{
	"CREATING":  BackupSummaryLifecycleStateCreating,
	"ACTIVE":    BackupSummaryLifecycleStateActive,
	"DELETING":  BackupSummaryLifecycleStateDeleting,
	"DELETED":   BackupSummaryLifecycleStateDeleted,
	"FAILED":    BackupSummaryLifecycleStateFailed,
	"RESTORING": BackupSummaryLifecycleStateRestoring,
}

// GetBackupSummaryLifecycleStateEnumValues Enumerates the set of values for BackupSummaryLifecycleState
func GetBackupSummaryLifecycleStateEnumValues() []BackupSummaryLifecycleStateEnum {
	values := make([]BackupSummaryLifecycleStateEnum, 0)
	for _, v := range mappingBackupSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

// BackupSummaryTypeEnum Enum with underlying type: string
type BackupSummaryTypeEnum string

// Set of constants representing the allowable values for BackupSummaryType
const (
	BackupSummaryTypeIncremental BackupSummaryTypeEnum = "INCREMENTAL"
	BackupSummaryTypeFull        BackupSummaryTypeEnum = "FULL"
)

var mappingBackupSummaryType = map[string]BackupSummaryTypeEnum{
	"INCREMENTAL": BackupSummaryTypeIncremental,
	"FULL":        BackupSummaryTypeFull,
}

// GetBackupSummaryTypeEnumValues Enumerates the set of values for BackupSummaryType
func GetBackupSummaryTypeEnumValues() []BackupSummaryTypeEnum {
	values := make([]BackupSummaryTypeEnum, 0)
	for _, v := range mappingBackupSummaryType {
		values = append(values, v)
	}
	return values
}

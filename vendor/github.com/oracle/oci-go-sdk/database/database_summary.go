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

// DatabaseSummary An Oracle database on a DB System. For more information, see Managing Oracle Databases (https://docs.us-phoenix-1.oraclecloud.com/Content/Database/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an administrator. If you're an administrator who needs to write policies to give users access, see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type DatabaseSummary struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The database name.
	DbName *string `mandatory:"true" json:"dbName"`

	// A system-generated name for the database to ensure uniqueness within an Oracle Data Guard group (a primary database and its standby databases). The unique name cannot be changed.
	DbUniqueName *string `mandatory:"true" json:"dbUniqueName"`

	// The OCID of the database.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the database.
	LifecycleState DatabaseSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The character set for the database.
	CharacterSet *string `mandatory:"false" json:"characterSet"`

	DbBackupConfig *DbBackupConfig `mandatory:"false" json:"dbBackupConfig"`

	// The OCID of the database home.
	DbHomeId *string `mandatory:"false" json:"dbHomeId"`

	// Database workload type.
	DbWorkload *string `mandatory:"false" json:"dbWorkload"`

	// Additional information about the current lifecycleState.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The national character set for the database.
	NcharacterSet *string `mandatory:"false" json:"ncharacterSet"`

	// Pluggable database name. It must begin with an alphabetic character and can contain a maximum of eight alphanumeric characters. Special characters are not permitted. Pluggable database should not be same as database name.
	PdbName *string `mandatory:"false" json:"pdbName"`

	// The date and time the database was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m DatabaseSummary) String() string {
	return common.PointerString(m)
}

// DatabaseSummaryLifecycleStateEnum Enum with underlying type: string
type DatabaseSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for DatabaseSummaryLifecycleState
const (
	DatabaseSummaryLifecycleStateProvisioning     DatabaseSummaryLifecycleStateEnum = "PROVISIONING"
	DatabaseSummaryLifecycleStateAvailable        DatabaseSummaryLifecycleStateEnum = "AVAILABLE"
	DatabaseSummaryLifecycleStateUpdating         DatabaseSummaryLifecycleStateEnum = "UPDATING"
	DatabaseSummaryLifecycleStateBackupInProgress DatabaseSummaryLifecycleStateEnum = "BACKUP_IN_PROGRESS"
	DatabaseSummaryLifecycleStateTerminating      DatabaseSummaryLifecycleStateEnum = "TERMINATING"
	DatabaseSummaryLifecycleStateTerminated       DatabaseSummaryLifecycleStateEnum = "TERMINATED"
	DatabaseSummaryLifecycleStateRestoreFailed    DatabaseSummaryLifecycleStateEnum = "RESTORE_FAILED"
	DatabaseSummaryLifecycleStateFailed           DatabaseSummaryLifecycleStateEnum = "FAILED"
)

var mappingDatabaseSummaryLifecycleState = map[string]DatabaseSummaryLifecycleStateEnum{
	"PROVISIONING":       DatabaseSummaryLifecycleStateProvisioning,
	"AVAILABLE":          DatabaseSummaryLifecycleStateAvailable,
	"UPDATING":           DatabaseSummaryLifecycleStateUpdating,
	"BACKUP_IN_PROGRESS": DatabaseSummaryLifecycleStateBackupInProgress,
	"TERMINATING":        DatabaseSummaryLifecycleStateTerminating,
	"TERMINATED":         DatabaseSummaryLifecycleStateTerminated,
	"RESTORE_FAILED":     DatabaseSummaryLifecycleStateRestoreFailed,
	"FAILED":             DatabaseSummaryLifecycleStateFailed,
}

// GetDatabaseSummaryLifecycleStateEnumValues Enumerates the set of values for DatabaseSummaryLifecycleState
func GetDatabaseSummaryLifecycleStateEnumValues() []DatabaseSummaryLifecycleStateEnum {
	values := make([]DatabaseSummaryLifecycleStateEnum, 0)
	for _, v := range mappingDatabaseSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

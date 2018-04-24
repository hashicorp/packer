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

// DbHomeSummary A directory where Oracle database software is installed. Each DB System can have multiple database homes,
// and each database home can have multiple databases within it. All the databases within a single database home
// must be the same database version, but different database homes can run different versions. For more information,
// see Managing Oracle Databases (https://docs.us-phoenix-1.oraclecloud.com/Content/Database/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an
// administrator. If you're an administrator who needs to write policies to give users access,
// see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type DbHomeSummary struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The Oracle database version.
	DbVersion *string `mandatory:"true" json:"dbVersion"`

	// The user-provided name for the database home. It does not need to be unique.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the database home.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the database home.
	LifecycleState DbHomeSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID of the DB System.
	DbSystemId *string `mandatory:"false" json:"dbSystemId"`

	// The OCID of the last patch history. This is updated as soon as a patch operation is started.
	LastPatchHistoryEntryId *string `mandatory:"false" json:"lastPatchHistoryEntryId"`

	// The date and time the database home was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m DbHomeSummary) String() string {
	return common.PointerString(m)
}

// DbHomeSummaryLifecycleStateEnum Enum with underlying type: string
type DbHomeSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for DbHomeSummaryLifecycleState
const (
	DbHomeSummaryLifecycleStateProvisioning DbHomeSummaryLifecycleStateEnum = "PROVISIONING"
	DbHomeSummaryLifecycleStateAvailable    DbHomeSummaryLifecycleStateEnum = "AVAILABLE"
	DbHomeSummaryLifecycleStateUpdating     DbHomeSummaryLifecycleStateEnum = "UPDATING"
	DbHomeSummaryLifecycleStateTerminating  DbHomeSummaryLifecycleStateEnum = "TERMINATING"
	DbHomeSummaryLifecycleStateTerminated   DbHomeSummaryLifecycleStateEnum = "TERMINATED"
	DbHomeSummaryLifecycleStateFailed       DbHomeSummaryLifecycleStateEnum = "FAILED"
)

var mappingDbHomeSummaryLifecycleState = map[string]DbHomeSummaryLifecycleStateEnum{
	"PROVISIONING": DbHomeSummaryLifecycleStateProvisioning,
	"AVAILABLE":    DbHomeSummaryLifecycleStateAvailable,
	"UPDATING":     DbHomeSummaryLifecycleStateUpdating,
	"TERMINATING":  DbHomeSummaryLifecycleStateTerminating,
	"TERMINATED":   DbHomeSummaryLifecycleStateTerminated,
	"FAILED":       DbHomeSummaryLifecycleStateFailed,
}

// GetDbHomeSummaryLifecycleStateEnumValues Enumerates the set of values for DbHomeSummaryLifecycleState
func GetDbHomeSummaryLifecycleStateEnumValues() []DbHomeSummaryLifecycleStateEnum {
	values := make([]DbHomeSummaryLifecycleStateEnum, 0)
	for _, v := range mappingDbHomeSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

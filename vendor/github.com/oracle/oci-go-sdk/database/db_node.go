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

// DbNode A server where Oracle database software is running.
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an administrator. If you're an administrator who needs to write policies to give users access, see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type DbNode struct {

	// The OCID of the DB System.
	DbSystemId *string `mandatory:"true" json:"dbSystemId"`

	// The OCID of the DB Node.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the database node.
	LifecycleState DbNodeLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time that the DB Node was created.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the VNIC.
	VnicId *string `mandatory:"true" json:"vnicId"`

	// The OCID of the backup VNIC.
	BackupVnicId *string `mandatory:"false" json:"backupVnicId"`

	// The host name for the DB Node.
	Hostname *string `mandatory:"false" json:"hostname"`

	// Storage size, in GBs, of the software volume that is allocated to the DB system. This is applicable only for VM-based DBs.
	SoftwareStorageSizeInGB *int `mandatory:"false" json:"softwareStorageSizeInGB"`
}

func (m DbNode) String() string {
	return common.PointerString(m)
}

// DbNodeLifecycleStateEnum Enum with underlying type: string
type DbNodeLifecycleStateEnum string

// Set of constants representing the allowable values for DbNodeLifecycleState
const (
	DbNodeLifecycleStateProvisioning DbNodeLifecycleStateEnum = "PROVISIONING"
	DbNodeLifecycleStateAvailable    DbNodeLifecycleStateEnum = "AVAILABLE"
	DbNodeLifecycleStateUpdating     DbNodeLifecycleStateEnum = "UPDATING"
	DbNodeLifecycleStateStopping     DbNodeLifecycleStateEnum = "STOPPING"
	DbNodeLifecycleStateStopped      DbNodeLifecycleStateEnum = "STOPPED"
	DbNodeLifecycleStateStarting     DbNodeLifecycleStateEnum = "STARTING"
	DbNodeLifecycleStateTerminating  DbNodeLifecycleStateEnum = "TERMINATING"
	DbNodeLifecycleStateTerminated   DbNodeLifecycleStateEnum = "TERMINATED"
	DbNodeLifecycleStateFailed       DbNodeLifecycleStateEnum = "FAILED"
)

var mappingDbNodeLifecycleState = map[string]DbNodeLifecycleStateEnum{
	"PROVISIONING": DbNodeLifecycleStateProvisioning,
	"AVAILABLE":    DbNodeLifecycleStateAvailable,
	"UPDATING":     DbNodeLifecycleStateUpdating,
	"STOPPING":     DbNodeLifecycleStateStopping,
	"STOPPED":      DbNodeLifecycleStateStopped,
	"STARTING":     DbNodeLifecycleStateStarting,
	"TERMINATING":  DbNodeLifecycleStateTerminating,
	"TERMINATED":   DbNodeLifecycleStateTerminated,
	"FAILED":       DbNodeLifecycleStateFailed,
}

// GetDbNodeLifecycleStateEnumValues Enumerates the set of values for DbNodeLifecycleState
func GetDbNodeLifecycleStateEnumValues() []DbNodeLifecycleStateEnum {
	values := make([]DbNodeLifecycleStateEnum, 0)
	for _, v := range mappingDbNodeLifecycleState {
		values = append(values, v)
	}
	return values
}

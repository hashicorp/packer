// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
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
	"github.com/oracle/oci-go-sdk/v36/common"
)

// InstanceConsoleConnection The `InstanceConsoleConnection` API provides you with console access to Compute instances,
// enabling you to troubleshoot malfunctioning instances remotely.
// For more information about instance console connections, see Troubleshooting Instances Using Instance Console Connections (https://docs.cloud.oracle.com/Content/Compute/References/serialconsole.htm).
type InstanceConsoleConnection struct {

	// The OCID of the compartment to contain the console connection.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The SSH connection string for the console connection.
	ConnectionString *string `mandatory:"false" json:"connectionString"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The SSH public key fingerprint for the console connection.
	Fingerprint *string `mandatory:"false" json:"fingerprint"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID of the console connection.
	Id *string `mandatory:"false" json:"id"`

	// The OCID of the instance the console connection connects to.
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// The current state of the console connection.
	LifecycleState InstanceConsoleConnectionLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The SSH connection string for the SSH tunnel used to
	// connect to the console connection over VNC.
	VncConnectionString *string `mandatory:"false" json:"vncConnectionString"`
}

func (m InstanceConsoleConnection) String() string {
	return common.PointerString(m)
}

// InstanceConsoleConnectionLifecycleStateEnum Enum with underlying type: string
type InstanceConsoleConnectionLifecycleStateEnum string

// Set of constants representing the allowable values for InstanceConsoleConnectionLifecycleStateEnum
const (
	InstanceConsoleConnectionLifecycleStateActive   InstanceConsoleConnectionLifecycleStateEnum = "ACTIVE"
	InstanceConsoleConnectionLifecycleStateCreating InstanceConsoleConnectionLifecycleStateEnum = "CREATING"
	InstanceConsoleConnectionLifecycleStateDeleted  InstanceConsoleConnectionLifecycleStateEnum = "DELETED"
	InstanceConsoleConnectionLifecycleStateDeleting InstanceConsoleConnectionLifecycleStateEnum = "DELETING"
	InstanceConsoleConnectionLifecycleStateFailed   InstanceConsoleConnectionLifecycleStateEnum = "FAILED"
)

var mappingInstanceConsoleConnectionLifecycleState = map[string]InstanceConsoleConnectionLifecycleStateEnum{
	"ACTIVE":   InstanceConsoleConnectionLifecycleStateActive,
	"CREATING": InstanceConsoleConnectionLifecycleStateCreating,
	"DELETED":  InstanceConsoleConnectionLifecycleStateDeleted,
	"DELETING": InstanceConsoleConnectionLifecycleStateDeleting,
	"FAILED":   InstanceConsoleConnectionLifecycleStateFailed,
}

// GetInstanceConsoleConnectionLifecycleStateEnumValues Enumerates the set of values for InstanceConsoleConnectionLifecycleStateEnum
func GetInstanceConsoleConnectionLifecycleStateEnumValues() []InstanceConsoleConnectionLifecycleStateEnum {
	values := make([]InstanceConsoleConnectionLifecycleStateEnum, 0)
	for _, v := range mappingInstanceConsoleConnectionLifecycleState {
		values = append(values, v)
	}
	return values
}

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

// InstanceConfigurationAvailabilityConfig Options for customers to define the general policy of how compute service perform maintenance on VM instances.
type InstanceConfigurationAvailabilityConfig struct {

	// Actions customers can specify that would be applied to their instances after scheduled or unexpected host maintenance.
	// * `RESTORE_INSTANCE` - This would be the default action if recoveryAction is not set. VM instances
	// will be restored to the power state it was in before maintenance.
	// * `STOP_INSTANCE` - This action allow customers to have their VM instances be stopped after maintenance.
	RecoveryAction InstanceConfigurationAvailabilityConfigRecoveryActionEnum `mandatory:"false" json:"recoveryAction,omitempty"`
}

func (m InstanceConfigurationAvailabilityConfig) String() string {
	return common.PointerString(m)
}

// InstanceConfigurationAvailabilityConfigRecoveryActionEnum Enum with underlying type: string
type InstanceConfigurationAvailabilityConfigRecoveryActionEnum string

// Set of constants representing the allowable values for InstanceConfigurationAvailabilityConfigRecoveryActionEnum
const (
	InstanceConfigurationAvailabilityConfigRecoveryActionRestoreInstance InstanceConfigurationAvailabilityConfigRecoveryActionEnum = "RESTORE_INSTANCE"
	InstanceConfigurationAvailabilityConfigRecoveryActionStopInstance    InstanceConfigurationAvailabilityConfigRecoveryActionEnum = "STOP_INSTANCE"
)

var mappingInstanceConfigurationAvailabilityConfigRecoveryAction = map[string]InstanceConfigurationAvailabilityConfigRecoveryActionEnum{
	"RESTORE_INSTANCE": InstanceConfigurationAvailabilityConfigRecoveryActionRestoreInstance,
	"STOP_INSTANCE":    InstanceConfigurationAvailabilityConfigRecoveryActionStopInstance,
}

// GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumValues Enumerates the set of values for InstanceConfigurationAvailabilityConfigRecoveryActionEnum
func GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumValues() []InstanceConfigurationAvailabilityConfigRecoveryActionEnum {
	values := make([]InstanceConfigurationAvailabilityConfigRecoveryActionEnum, 0)
	for _, v := range mappingInstanceConfigurationAvailabilityConfigRecoveryAction {
		values = append(values, v)
	}
	return values
}

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

// InstanceConfigurationAvailabilityConfig Options for defining the availabiity of a VM instance after a maintenance event that impacts the underlying hardware.
type InstanceConfigurationAvailabilityConfig struct {

	// The lifecycle state for an instance when it is recovered after infrastructure maintenance.
	// * `RESTORE_INSTANCE` - The instance is restored to the lifecycle state it was in before the maintenance event.
	// If the instance was running, it is automatically rebooted. This is the default action when a value is not set.
	// * `STOP_INSTANCE` - The instance is recovered in the stopped state.
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

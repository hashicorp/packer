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

// InstanceAgentPluginConfigDetails The configuration of plugins associated with this instance.
type InstanceAgentPluginConfigDetails struct {

	// The plugin name. To get a list of available plugins, use the
	// ListInstanceagentAvailablePlugins
	// operation in the Oracle Cloud Agent API. For more information about the available plugins, see
	// Managing Plugins with Oracle Cloud Agent (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/manage-plugins.htm).
	Name *string `mandatory:"true" json:"name"`

	// Whether the plugin should be enabled or disabled.
	// To enable the monitoring and management plugins, the `isMonitoringDisabled` and
	// `isManagementDisabled` attributes must also be set to false.
	DesiredState InstanceAgentPluginConfigDetailsDesiredStateEnum `mandatory:"true" json:"desiredState"`
}

func (m InstanceAgentPluginConfigDetails) String() string {
	return common.PointerString(m)
}

// InstanceAgentPluginConfigDetailsDesiredStateEnum Enum with underlying type: string
type InstanceAgentPluginConfigDetailsDesiredStateEnum string

// Set of constants representing the allowable values for InstanceAgentPluginConfigDetailsDesiredStateEnum
const (
	InstanceAgentPluginConfigDetailsDesiredStateEnabled  InstanceAgentPluginConfigDetailsDesiredStateEnum = "ENABLED"
	InstanceAgentPluginConfigDetailsDesiredStateDisabled InstanceAgentPluginConfigDetailsDesiredStateEnum = "DISABLED"
)

var mappingInstanceAgentPluginConfigDetailsDesiredState = map[string]InstanceAgentPluginConfigDetailsDesiredStateEnum{
	"ENABLED":  InstanceAgentPluginConfigDetailsDesiredStateEnabled,
	"DISABLED": InstanceAgentPluginConfigDetailsDesiredStateDisabled,
}

// GetInstanceAgentPluginConfigDetailsDesiredStateEnumValues Enumerates the set of values for InstanceAgentPluginConfigDetailsDesiredStateEnum
func GetInstanceAgentPluginConfigDetailsDesiredStateEnumValues() []InstanceAgentPluginConfigDetailsDesiredStateEnum {
	values := make([]InstanceAgentPluginConfigDetailsDesiredStateEnum, 0)
	for _, v := range mappingInstanceAgentPluginConfigDetailsDesiredState {
		values = append(values, v)
	}
	return values
}

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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v36/common"
)

// InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig The platform configuration used when launching a bare metal instance specific to the AMD Milan platform.
type InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig struct {

	// The number of NUMA nodes per socket.
	NumaNodesPerSocket InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum `mandatory:"false" json:"numaNodesPerSocket,omitempty"`
}

func (m InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig
	}{
		"AMD_MILAN_BM",
		(MarshalTypeInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfig)(m),
	}

	return json.Marshal(&s)
}

// InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum Enum with underlying type: string
type InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum string

// Set of constants representing the allowable values for InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum
const (
	InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps0 InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS0"
	InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps1 InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS1"
	InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps2 InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS2"
	InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps4 InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS4"
)

var mappingInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocket = map[string]InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum{
	"NPS0": InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps0,
	"NPS1": InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps1,
	"NPS2": InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps2,
	"NPS4": InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketNps4,
}

// GetInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnumValues Enumerates the set of values for InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum
func GetInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnumValues() []InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum {
	values := make([]InstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocketEnum, 0)
	for _, v := range mappingInstanceConfigurationAmdMilanBmLaunchInstancePlatformConfigNumaNodesPerSocket {
		values = append(values, v)
	}
	return values
}

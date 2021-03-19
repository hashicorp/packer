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

// AmdMilanBmPlatformConfig The platform configuration of a bare metal instance specific to the AMD Milan platform.
type AmdMilanBmPlatformConfig struct {

	// The number of NUMA nodes per socket.
	NumaNodesPerSocket AmdMilanBmPlatformConfigNumaNodesPerSocketEnum `mandatory:"false" json:"numaNodesPerSocket,omitempty"`
}

func (m AmdMilanBmPlatformConfig) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m AmdMilanBmPlatformConfig) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeAmdMilanBmPlatformConfig AmdMilanBmPlatformConfig
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeAmdMilanBmPlatformConfig
	}{
		"AMD_MILAN_BM",
		(MarshalTypeAmdMilanBmPlatformConfig)(m),
	}

	return json.Marshal(&s)
}

// AmdMilanBmPlatformConfigNumaNodesPerSocketEnum Enum with underlying type: string
type AmdMilanBmPlatformConfigNumaNodesPerSocketEnum string

// Set of constants representing the allowable values for AmdMilanBmPlatformConfigNumaNodesPerSocketEnum
const (
	AmdMilanBmPlatformConfigNumaNodesPerSocketNps0 AmdMilanBmPlatformConfigNumaNodesPerSocketEnum = "NPS0"
	AmdMilanBmPlatformConfigNumaNodesPerSocketNps1 AmdMilanBmPlatformConfigNumaNodesPerSocketEnum = "NPS1"
	AmdMilanBmPlatformConfigNumaNodesPerSocketNps2 AmdMilanBmPlatformConfigNumaNodesPerSocketEnum = "NPS2"
	AmdMilanBmPlatformConfigNumaNodesPerSocketNps4 AmdMilanBmPlatformConfigNumaNodesPerSocketEnum = "NPS4"
)

var mappingAmdMilanBmPlatformConfigNumaNodesPerSocket = map[string]AmdMilanBmPlatformConfigNumaNodesPerSocketEnum{
	"NPS0": AmdMilanBmPlatformConfigNumaNodesPerSocketNps0,
	"NPS1": AmdMilanBmPlatformConfigNumaNodesPerSocketNps1,
	"NPS2": AmdMilanBmPlatformConfigNumaNodesPerSocketNps2,
	"NPS4": AmdMilanBmPlatformConfigNumaNodesPerSocketNps4,
}

// GetAmdMilanBmPlatformConfigNumaNodesPerSocketEnumValues Enumerates the set of values for AmdMilanBmPlatformConfigNumaNodesPerSocketEnum
func GetAmdMilanBmPlatformConfigNumaNodesPerSocketEnumValues() []AmdMilanBmPlatformConfigNumaNodesPerSocketEnum {
	values := make([]AmdMilanBmPlatformConfigNumaNodesPerSocketEnum, 0)
	for _, v := range mappingAmdMilanBmPlatformConfigNumaNodesPerSocket {
		values = append(values, v)
	}
	return values
}

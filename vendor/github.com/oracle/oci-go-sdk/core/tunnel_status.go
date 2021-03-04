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

// TunnelStatus Deprecated. For tunnel information, instead see IPSecConnectionTunnel.
type TunnelStatus struct {

	// The IP address of Oracle's VPN headend.
	// Example: `203.0.113.50`
	IpAddress *string `mandatory:"true" json:"ipAddress"`

	// The tunnel's current state.
	LifecycleState TunnelStatusLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the IPSec connection was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// When the state of the tunnel last changed, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeStateModified *common.SDKTime `mandatory:"false" json:"timeStateModified"`
}

func (m TunnelStatus) String() string {
	return common.PointerString(m)
}

// TunnelStatusLifecycleStateEnum Enum with underlying type: string
type TunnelStatusLifecycleStateEnum string

// Set of constants representing the allowable values for TunnelStatusLifecycleStateEnum
const (
	TunnelStatusLifecycleStateUp                 TunnelStatusLifecycleStateEnum = "UP"
	TunnelStatusLifecycleStateDown               TunnelStatusLifecycleStateEnum = "DOWN"
	TunnelStatusLifecycleStateDownForMaintenance TunnelStatusLifecycleStateEnum = "DOWN_FOR_MAINTENANCE"
)

var mappingTunnelStatusLifecycleState = map[string]TunnelStatusLifecycleStateEnum{
	"UP":                   TunnelStatusLifecycleStateUp,
	"DOWN":                 TunnelStatusLifecycleStateDown,
	"DOWN_FOR_MAINTENANCE": TunnelStatusLifecycleStateDownForMaintenance,
}

// GetTunnelStatusLifecycleStateEnumValues Enumerates the set of values for TunnelStatusLifecycleStateEnum
func GetTunnelStatusLifecycleStateEnumValues() []TunnelStatusLifecycleStateEnum {
	values := make([]TunnelStatusLifecycleStateEnum, 0)
	for _, v := range mappingTunnelStatusLifecycleState {
		values = append(values, v)
	}
	return values
}

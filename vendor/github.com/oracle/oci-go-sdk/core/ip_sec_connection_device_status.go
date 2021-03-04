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

// IpSecConnectionDeviceStatus Deprecated. For tunnel information, instead see
// IPSecConnectionTunnel.
type IpSecConnectionDeviceStatus struct {

	// The OCID of the compartment containing the IPSec connection.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The IPSec connection's Oracle ID (OCID).
	Id *string `mandatory:"true" json:"id"`

	// The date and time the IPSec connection was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Two TunnelStatus objects.
	Tunnels []TunnelStatus `mandatory:"false" json:"tunnels"`
}

func (m IpSecConnectionDeviceStatus) String() string {
	return common.PointerString(m)
}
